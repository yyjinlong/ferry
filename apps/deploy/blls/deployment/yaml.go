// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"fmt"

	"ferry/ops/log"
)

type yaml struct {
	pipelineID  int64  // 流水线ID
	phase       string // 部署阶段
	deployment  string // deployment名字
	namespace   string // 当前服务所在命名空间
	service     string // 服务名
	imageURL    string // 镜像地址
	imageTag    string // 镜像tag
	replicas    int    // 副本数
	quotaCpu    int    // cpu配额
	quotaMaxCpu int    // cpu最大配额
	quotaMem    int    // 内存配额
	quotaMaxMem int    // 内存最大配额
	volumeConf  string // 数据卷配置
	reserveTime int    // 终止后的预留时间
}

func (y *yaml) instance() (string, error) {
	/*
		apiVersion
		kind
		metadata
		  ...
		spec:
		  ...
	*/
	guest := make(map[string]interface{})
	guest["apiVersion"] = "apps/v1"
	guest["kind"] = "Deployment"
	guest["metadata"] = y.metadata()

	spec, err := y.spec()
	if err != nil {
		return "", err
	}
	guest["spec"] = spec

	config, err := json.Marshal(guest)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (y *yaml) metadata() map[string]string {
	meta := make(map[string]string)
	meta["name"] = y.deployment
	meta["namespace"] = y.namespace
	return meta
}

func (y *yaml) spec() (map[string]interface{}, error) {
	/*
		spec:
		  replicas:
		  selector:
		    ...
		  strategy:
		    ...
		  template:
		    ...
	*/
	spec := make(map[string]interface{})
	spec["replicas"] = y.replicas
	spec["selector"] = y.selector()
	spec["strategy"] = y.strategy()

	template, err := y.template()
	if err != nil {
		return nil, err
	}
	spec["template"] = template
	return spec, nil
}

func (y *yaml) selector() map[string]interface{} {
	selector := map[string]interface{}{
		"matchLabels": y.labels(),
	}
	return selector
}

func (y *yaml) labels() map[string]string {
	return map[string]string{
		"service": y.service,
		"phase":   y.phase,
	}
}

func (y *yaml) strategy() map[string]interface{} {
	rollingUpdate := map[string]interface{}{
		"maxSurge":       0,
		"maxUnavailable": "100%",
	}
	return map[string]interface{}{
		"type":          "RollingUpdate",
		"rollingUpdate": rollingUpdate,
	}
}

func (y *yaml) template() (map[string]interface{}, error) {
	/*
		template:
		  metadata:
		    ...
		  spec:
		    ...
	*/
	tpl := make(map[string]interface{})
	tpl["metadata"] = y.templateMetadata()

	spec, err := y.templateSpec()
	if err != nil {
		return nil, err
	}
	tpl["spec"] = spec
	return tpl, nil
}

func (y *yaml) templateMetadata() interface{} {
	labels := make(map[string]interface{})
	labels["labels"] = y.labels()
	return labels
}

func (y *yaml) templateSpec() (interface{}, error) {
	/*
		spec:
		  hostAliases:
			...
		  dnsConfig:
			...
		  dnsPolicy:
		  terminationGracePeriodSeconds:
		  volumes:
			...
		  initContainers:
			...
		  containers:
			...
		  affinity:
			...
	*/
	spec := make(map[string]interface{})
	spec["hostAliases"] = y.hostAliases()
	spec["dnsConfig"] = y.dnsConfig()
	spec["dnsPolicy"] = "None"
	spec["terminationGracePeriodSeconds"] = y.reserveTime

	volumes, err := y.volumes()
	if err != nil {
		return nil, err
	}
	spec["volumes"] = volumes

	containers, err := y.containers()
	if err != nil {
		return nil, err
	}
	spec["containers"] = containers
	return spec, nil
}

func (y *yaml) hostAliases() interface{} {
	/*
	   hostAliases:
	     - hostnames:
	         - ...
	       ip:
	*/
	// 默认主机配置
	hosts := []map[string]string{{"127.0.0.1": "localhost.localdomain"}}

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
		hostAliases := make(map[string]interface{})
		hostAliases["ip"] = ip
		hostAliases["hostnames"] = hostList
		hostAliaseList = append(hostAliaseList, hostAliases)
	}
	return hostAliaseList
}

func (y *yaml) dnsConfig() interface{} {
	/*
		dnsConfig:
			nameservers:
			- ...
	*/
	dnsList := []string{"114.114.114.114"}
	return map[string][]string{
		"nameservers": dnsList,
	}
}

func (y *yaml) volumes() (interface{}, error) {
	// NOTE: 在宿主机上创建本地存储卷, 目前只支持hostPath类型.
	volumes := make([]interface{}, 0)
	defineVolume, err := y.createDefineVolume()
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, defineVolume)
	return volumes, nil
}

func (y *yaml) createDefineVolume() (interface{}, error) {
	/*
	   创建自定义的数据卷(服务需要的数据卷)
	   volumes:
	     - name:
	       hostPath:
	         path:
	         type:
	*/

	type physicalInfo struct {
		HostpathType string `json:"hostpath_type"`
		PhysicalPath string `json:"physical_path"`
	}

	type volume struct {
		Type     string       `json:"volume_type"`
		Name     string       `json:"volume_name"`
		Physical physicalInfo `json:"physical"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(y.volumeConf), &volumes); err != nil {
		return nil, err
	}

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
	return defineVolume, nil
}

func (y *yaml) containers() (interface{}, error) {
	/*
	   - name:
	     image:
	     imagePullPolicy:
	     env:
	     lifecycle:
	       ...
	     resources:
	       ...
	     securityContext:
	       ...
	     volumeMounts:
	       ...
	     livenessProbe:
	       ...
	     readinessProbe:
	       ...
	     terminationMessagePath:
	     terminationMessagePolicy:
	*/

	containerList := make([]interface{}, 0)
	container := make(map[string]interface{})
	container["name"] = y.service
	container["image"] = fmt.Sprintf("%s:%s", y.imageURL, y.imageTag)
	container["imagePullPolicy"] = "IfNotPresent"
	container["env"] = y.setEnv()
	container["resources"] = y.setResource(y.quotaCpu, y.quotaMaxCpu, y.quotaMem, y.quotaMaxMem)
	container["securityContext"] = y.security()
	container["lifecycle"] = y.lifecycle()
	volumeMounts, err := y.mountContainerVolume()
	if err != nil {
		return nil, err
	}
	container["volumeMounts"] = volumeMounts
	container["livenessProbe"] = y.liveness()
	container["readinessProbe"] = y.readiness()
	container["terminationMessagePath"] = "/dev/termination-log"
	container["terminationMessagePolicy"] = "File"
	containerList = append(containerList, container)
	return containerList, nil
}

func (y *yaml) setEnv() interface{} {
	/*
	   - env:
	       - name:
	         value:
	*/
	env := []map[string]string{
		{"name": "SERVICE", "value": y.service},
	}
	return env
}

func (y *yaml) setResource(cpu, cpuMax, mem, memMax int) interface{} {
	/*
		resources:
		    requests:
			  ...
		    limits:
			  ...
	*/
	requests := map[string]string{
		"cpu":    fmt.Sprintf("%dm", cpu),
		"memory": fmt.Sprintf("%dMi", mem),
	}
	limits := map[string]string{
		"cpu":    fmt.Sprintf("%dm", cpuMax),
		"memory": fmt.Sprintf("%dMi", memMax),
	}
	return map[string]interface{}{
		"requests": requests,
		"limits":   limits,
	}
}

func (y *yaml) security() interface{} {
	/*
	   securityContext:
	     capabilities:
	       add:
	*/

	sysList := []string{"SYS_ADMIN", "SYS_PTRACE"}
	capabilities := map[string][]string{"add": sysList}
	context := map[string]interface{}{
		"capabilities": capabilities,
	}
	return context
}

func (y *yaml) lifecycle() interface{} {
	/*
	   在容器被终结之前, Kubernetes 将发送一个 preStop 事件.
	   优雅关闭: 先发送一个kill信号(kill -3), 之后sleep 30秒等待未处理完的请求,
	             如果没处理完, 则会被强制终止(kill -9)
	   lifecycle:
	     preStop:
	       exec:
	         command:
	*/
	stopCmd := []string{
		"/bin/sh",
		"-c",
		"sleep 30",
	}
	stopExec := map[string]interface{}{"command": stopCmd}
	preStop := map[string]interface{}{"exec": stopExec}

	life := make(map[string]interface{})
	life["preStop"] = preStop
	return life
}

func (y *yaml) mountContainerVolume() (interface{}, error) {
	/*
	   volumeMounts:
	     - mountPath:
	       name:
	       subPath:
	*/

	type containerInfo struct {
		ContainerPath string `json:"container_path"`
	}

	type volume struct {
		Type      string        `json:"volume_type"`
		Name      string        `json:"volume_name"`
		Container containerInfo `json:"container"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(y.volumeConf), &volumes); err != nil {
		return nil, err
	}

	containerVolumeMounts := make([]map[string]string, 0)
	for _, item := range volumes {
		containerVolume := map[string]string{
			"name":      item.Name,
			"mountPath": item.Container.ContainerPath,
		}
		containerVolumeMounts = append(containerVolumeMounts, containerVolume)
		log.Infof("container volume mount path: %s to: %s", item.Container.ContainerPath, item.Name)
	}
	return containerVolumeMounts, nil
}

func (y *yaml) liveness() interface{} {
	/*
		exec:
		  command:
		    ...
		initialDelaySeconds:
		timeoutSeconds:
		periodSeconds:
		successThreshold:
		failureThreshold:
	*/

	command := []string{
		"/bin/sh",
		"/home/tong/opbin/liveness-prob.sh",
	}
	exec := map[string][]string{"command": command}

	live := make(map[string]interface{})
	live["exec"] = exec
	live["initialDelaySeconds"] = 5
	live["timeoutSeconds"] = 5
	live["periodSeconds"] = 60
	live["successThreshold"] = 1
	live["failureThreshold"] = 3
	return live
}

func (y *yaml) readiness() interface{} {
	/*
		exec:
		  command:
		    ...
		initialDelaySeconds:
		timeoutSeconds:
		periodSeconds:
		successThreshold:
		failureThreshold:
	*/

	command := []string{
		"/bin/sh",
		"/home/tong/opbin/readiness-probe.sh",
	}
	exec := map[string][]string{"command": command}

	ready := make(map[string]interface{})
	ready["exec"] = exec
	ready["initialDelaySeconds"] = 5
	ready["timeoutSeconds"] = 10
	ready["periodSeconds"] = 10
	ready["successThreshold"] = 1
	ready["failureThreshold"] = 10
	return ready
}
