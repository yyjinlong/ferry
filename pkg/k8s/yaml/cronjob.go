package yaml

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	CRON_BOOT_DEADLINE          = 180
	CRON_PHASE                  = "cronjob"
	CRON_TERMINATE_MESSAGE_PATH = "/dev/termination-log"
	CRON_TERMINATE_POLICY       = "File"
)

type CronjobYaml struct {
	Namespace   string
	Service     string
	ImageURL    string
	ImageTag    string
	VolumeConf  string // 数据卷配置
	ReserveTime int    // 终止后的预留时间
	Name        string // cronjob的名字
	Schedule    string // cronjob的调度时间
	Command     string // cronjob执行的命令
}

func (cy *CronjobYaml) Instance() (string, error) {
	controller := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "CronJob",
		"metadata":   cy.cronMetadata(),
	}

	spec, err := cy.cronSpec()
	if err != nil {
		return "", err
	}
	controller["spec"] = spec

	config, err := json.Marshal(controller)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (cy *CronjobYaml) cronMetadata() map[string]string {
	return map[string]string{
		"name":      cy.Name,
		"namespace": cy.Namespace,
	}
}

func (cy *CronjobYaml) cronSpec() (interface{}, error) {
	/*
	  spec:
	    schedule:
	    concurrencyPolicy:
	    startingDeadlineSeconds:
	    successfulJobsHistoryLimit:
	    failedJobsHistoryLimit:
	    suspend:

	    jobTemplate:
	*/
	spec := map[string]interface{}{
		"schedule":                   cy.Schedule,
		"concurrencyPolicy":          "Forbid",           // 类似文件锁
		"startingDeadlineSeconds":    CRON_BOOT_DEADLINE, // 开始该任务的截止时间秒数
		"successfulJobsHistoryLimit": 0,                  // 保留多少已完成的任务数
		"failedJobsHistoryLimit":     0,                  // 保留多少失败的任务数
		"suspend":                    false,
	}

	jobTpl, err := cy.jobTemplate()
	if err != nil {
		return nil, err
	}
	spec["jobTemplate"] = jobTpl
	return spec, nil
}

func (cy *CronjobYaml) jobTemplate() (interface{}, error) {
	/*
	  spec:
	    parallelism:
	    completions:
	    backoffLimit:

	    template:
	*/
	spec := map[string]interface{}{
		"parallelism":  1, // 并发启动pod数目
		"completions":  1, // 至少要完成的pod数目
		"backoffLimit": 0, // job的重试次数
	}

	podTpl, err := cy.podTemplate()
	if err != nil {
		return nil, err
	}
	spec["template"] = podTpl

	return map[string]interface{}{
		"spec": spec,
	}, nil
}

func (cy *CronjobYaml) podTemplate() (interface{}, error) {
	/*
	  template:
	    metadata:
	      labels:
	        ...
	    spec:
	      ...
	*/

	spec, err := cy.podSpec()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"metadata": cy.podMetaData(),
		"spec":     spec,
	}, nil
}

func (cy *CronjobYaml) podMetaData() interface{} {
	/*
	  labels:
	    ...
	*/
	labels := map[string]string{
		"service": cy.Service,
		"appid":   cy.Name,
		"phase":   CRON_PHASE,
	}
	return map[string]interface{}{
		"labels": labels,
	}
}

func (cy *CronjobYaml) podSpec() (interface{}, error) {
	/*
	  spec:
	    initContainers:
	      ...
	    containers:
	      ...
	    imagePullSecrets:
	      ...
	    nodeSelector:
	      ...
	    restartPolicy:Never:
	    volumes:
	      ...
	    hostAliases:
	      ...
	    dnsPolicy:
	      ...
	    dnsConfig:
	      ...
	    terminationGracePeriodSeconds:
	*/

	containers, err := cy.containers()
	if err != nil {
		return nil, err
	}

	spec := make(map[string]interface{})
	spec["containers"] = containers
	spec["imagePullSecrets"] = cy.imagePullSecrets()
	spec["nodeSelector"] = cy.nodeSelector()
	spec["restartPolicy"] = "Never"

	volumes, err := cy.volumes()
	if err != nil {
		return nil, err
	}
	spec["volumes"] = volumes

	spec["hostAliases"] = cy.hostAliases()
	spec["dnsPolicy"] = "None"
	spec["dnsConfig"] = cy.dnsConfig()
	spec["terminationGracePeriodSeconds"] = cy.ReserveTime
	return spec, nil
}

func (cy *CronjobYaml) imagePullSecrets() interface{} {
	/*
	  imagePullSecrets:
	  - name: xxx
	*/
	return []map[string]string{{"name": "harborkey"}}
}

func (cy *CronjobYaml) nodeSelector() interface{} {
	/*
	  nodeSelector:
	    ...
	*/
	return map[string]string{
		"batch": "cronjob",
	}
}

func (cy *CronjobYaml) volumes() (interface{}, error) {
	/*
	  volumes:
	  - name:
	    hostPath:
	      path:
	      type:
	*/

	// NOTE: 在宿主机上创建本地存储卷, 目前只支持hostPath类型.
	type physicalInfo struct {
		HostpathType string `json:"hostpath_type"`
		PhysicalPath string `json:"physical_path"`
	}

	type volume struct {
		Name     string       `json:"volume_name"`
		Type     string       `json:"volume_type"`
		Physical physicalInfo `json:"physical"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(cy.VolumeConf), &volumes); err != nil {
		return nil, err
	}

	// NOTE: 创建自定义的数据卷(服务需要的数据卷)
	defineVolume := make(map[string]interface{})
	for _, item := range volumes {
		defineVolume["name"] = item.Name
		if item.Type == "hostPath" {
			hostPath := map[string]string{
				"type": item.Physical.HostpathType,
				"path": item.Physical.PhysicalPath,
			}
			defineVolume["hostPath"] = hostPath
		}
	}
	log.Infof("create define volume: %v finish", defineVolume)

	return []map[string]interface{}{
		defineVolume,
	}, nil
}

func (cy *CronjobYaml) hostAliases() interface{} {
	/*
	  hostAliases:
	  - hostnames:
	    - ...
	    ip:
	*/

	// 默认主机配置
	hosts := []map[string]string{
		{"127.0.0.1": "localhost.localdomain"},
	}

	hostMap := make(map[string][]string)
	for _, item := range hosts {
		for ip, hostname := range item {
			if hostList, ok := hostMap[ip]; !ok {
				hostMap[ip] = []string{hostname}
			} else {
				hostList = append(hostList, hostname)
				hostMap[ip] = hostList
			}
		}
	}

	hostAliaseList := make([]interface{}, 0)
	for ip, hostList := range hostMap {
		hostAliaseList = append(hostAliaseList, map[string]interface{}{
			"hostnames": hostList,
			"ip":        ip,
		})
	}
	return hostAliaseList
}

func (cy *CronjobYaml) dnsConfig() interface{} {
	/*
	  dnsConfig:
	    nameservers:
	    - xxx.xxx.xxx.xxx
	*/
	dns := []string{
		"114.114.114.114",
	}
	return map[string][]string{
		"nameservers": dns,
	}
}

func (cy *CronjobYaml) containers() (interface{}, error) {
	/*
	  containers:
	  - name:
	    image:
	    imagePullPolicy:
	    args:
	      ...
	    env:
	    resources:
	      ...
	    securityContext:
	      ...
	    volumeMounts:
	      ...
	    terminationMessagePath:
	    terminationMessagePolicy:
	*/

	containerList := make([]interface{}, 0)
	container := make(map[string]interface{})
	container["name"] = cy.Service
	container["image"] = fmt.Sprintf("%s:%s", cy.ImageURL, cy.ImageTag)
	container["imagePullPolicy"] = "IfNotPresent"
	container["args"] = cy.args()
	container["env"] = cy.envs()
	container["resources"] = cy.resources()
	container["securityContext"] = cy.security()

	mounts, err := cy.volumeMounts()
	if err != nil {
		return nil, err
	}
	container["volumeMounts"] = mounts
	container["terminationMessagePath"] = CRON_TERMINATE_MESSAGE_PATH
	container["terminationMessagePolicy"] = CRON_TERMINATE_POLICY
	containerList = append(containerList, container)
	return containerList, nil
}

func (cy *CronjobYaml) args() interface{} {
	/*
	  args:
	  -
	*/

	cmd := fmt.Sprintf("su - tong -c \"%s && sleep 10\"", cy.Command)
	return []string{"/bin/sh", "-c", cmd}
}

func (cy *CronjobYaml) envs() interface{} {
	/*
	  env:
	  - name:
	    value:
	*/
	ipRef := map[string]string{
		"apiVersion": "v1",
		"fieldPath":  "status.podIP",
	}
	ipVal := map[string]interface{}{
		"fieldRef": ipRef,
	}

	nameRef := map[string]string{
		"apiVersion": "v1",
		"fieldPath":  "metadata.name",
	}
	nameVal := map[string]interface{}{
		"fieldRef": nameRef,
	}

	return []map[string]interface{}{
		{"name": "NAMESPACE", "value": cy.Namespace},
		{"name": "SERVICE", "value": cy.Service},
		{"name": "STAGE", "value": CRON_PHASE},
		{"name": "PODIP", "valueFrom": ipVal},
		{"name": "POD_NAME", "valueFrom": nameVal},
	}
}

func (cy *CronjobYaml) resources() interface{} {
	/*
	  resources:
	    requests:
	      ...
	    limits:
	      ....
	*/
	requests := map[string]string{
		"cpu":    "100m",
		"memory": "200Mi",
	}
	limits := map[string]string{
		"cpu":    "2048m",
		"memory": "4096Mi",
	}
	return map[string]interface{}{
		"requests": requests,
		"limits":   limits,
	}
}

func (cy *CronjobYaml) security() interface{} {
	/*
	  securityContext:
	    capabilities:
	      add:
	      - SYS_ADMIN
	      - SYS_PTRACE
	*/
	capabilities := map[string]interface{}{
		"add": []string{"SYS_ADMIN", "SYS_PTRACE"},
	}
	return map[string]interface{}{
		"capabilities": capabilities,
	}
}

func (cy *CronjobYaml) volumeMounts() (interface{}, error) {
	/*
	  volumeMounts:
	  - name:
	    mountPath:
	    subPath:
	*/

	type containerInfo struct {
		ContainerPath string `json:"container_path"`
	}

	type volume struct {
		Name      string        `json:"volume_name"`
		Type      string        `json:"volume_type"`
		Container containerInfo `json:"container"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(cy.VolumeConf), &volumes); err != nil {
		return nil, err
	}

	containerVolumeMounts := make([]map[string]string, 0)
	for _, item := range volumes {
		containerVolumeMounts = append(containerVolumeMounts, map[string]string{
			"name":      item.Name,
			"mountPath": item.Container.ContainerPath,
		})
		log.Infof("container volume mount path: %s to: %s", item.Container.ContainerPath, item.Name)
	}
	return containerVolumeMounts, nil
}
