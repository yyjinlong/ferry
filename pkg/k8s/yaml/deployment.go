// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package yaml

import (
	"encoding/json"
	"fmt"

	"nautilus/golib/log"
)

const (
	DEPLOY_TERMINATE_MESSAGE_PATH = "/dev/termination-log"
	DEPLOY_TERMINATE_POLICY       = "File"
)

type DeploymentYaml struct {
	Phase       string // 部署阶段
	Deployment  string // deployment名字
	AppID       string // 应用标识, 用于过滤
	Namespace   string // 当前服务所在命名空间
	Service     string // 服务名
	ImageURL    string // 镜像地址
	ImageTag    string // 镜像tag
	Replicas    int    // 副本数
	QuotaCpu    int    // cpu配额
	QuotaMaxCpu int    // cpu最大配额
	QuotaMem    int    // 内存配额
	QuotaMaxMem int    // 内存最大配额
	VolumeConf  string // 数据卷配置
	ConfigMap   string // configmap名字
	ReserveTime int    // 终止后的预留时间
}

func (dy *DeploymentYaml) Instance() (string, error) {
	/*
		apiVersion
		kind
		metadata
		  ...
		spec:
		  ...
	*/

	spec, err := dy.spec()
	if err != nil {
		return "", err
	}
	controller := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   dy.metadata(),
		"spec":       spec,
	}
	config, err := json.Marshal(controller)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (dy *DeploymentYaml) metadata() map[string]string {
	return map[string]string{
		"name":      dy.Deployment,
		"namespace": dy.Namespace,
	}
}

func (dy *DeploymentYaml) spec() (map[string]interface{}, error) {
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
	spec["replicas"] = dy.Replicas
	spec["selector"] = dy.selector()
	spec["strategy"] = dy.strategy()

	template, err := dy.podTemplate()
	if err != nil {
		return nil, err
	}
	spec["template"] = template
	return spec, nil
}

func (dy *DeploymentYaml) selector() map[string]interface{} {
	selector := map[string]interface{}{
		"matchLabels": dy.labels(),
	}
	return selector
}

func (dy *DeploymentYaml) labels() map[string]string {
	return map[string]string{
		"service": dy.Deployment,
		"phase":   dy.Phase,
		"appid":   dy.AppID,
	}
}

func (dy *DeploymentYaml) strategy() map[string]interface{} {
	rollingUpdate := map[string]interface{}{
		"maxSurge":       0,
		"maxUnavailable": "100%",
	}
	return map[string]interface{}{
		"type":          "RollingUpdate",
		"rollingUpdate": rollingUpdate,
	}
}

func (dy *DeploymentYaml) podTemplate() (map[string]interface{}, error) {
	/*
	  template:
	    metadata:
	      ...
	    spec:
	      ...
	*/
	spec, err := dy.podSpec()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"metadata": dy.podMetadata(),
		"spec":     spec,
	}, nil
}

func (dy *DeploymentYaml) podMetadata() interface{} {
	return map[string]interface{}{
		"labels": dy.labels(),
	}
}

func (dy *DeploymentYaml) podSpec() (interface{}, error) {
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
	    volumes:
	      ...
	    hostAliases:
	      ...
	    dnsPolicy:
	      ...
	    dnsConfig:
	      ...
	    affinity:
	      ...
	    terminationGracePeriodSeconds:
	*/

	containers, err := dy.containers()
	if err != nil {
		return nil, err
	}

	spec := make(map[string]interface{})
	spec["containers"] = containers
	spec["imagePullSecrets"] = dy.imagePullSecrets()
	spec["nodeSelector"] = dy.nodeSelector()

	volumes, err := dy.volumes()
	if err != nil {
		return nil, err
	}
	spec["volumes"] = volumes

	spec["hostAliases"] = dy.hostAliases()
	spec["dnsPolicy"] = "None"
	spec["dnsConfig"] = dy.dnsConfig()
	spec["terminationGracePeriodSeconds"] = dy.ReserveTime
	spec["affinity"] = dy.affinity()
	return spec, nil
}

func (dy *DeploymentYaml) hostAliases() interface{} {
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
		hostAliaseList = append(hostAliaseList, map[string]interface{}{
			"hostnames": hostList,
			"ip":        ip,
		})
	}
	return hostAliaseList
}

func (dy *DeploymentYaml) dnsConfig() interface{} {
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

func (dy *DeploymentYaml) imagePullSecrets() interface{} {
	/*
	  imagePullSecrets:
	  - name: xxx
	*/
	return []map[string]string{{"name": "harborkey"}}
}

func (dy *DeploymentYaml) nodeSelector() interface{} {
	/*
	  nodeSelector:
	    ...
	*/
	return map[string]string{
		"aggregate": "default",
	}
}

func (dy *DeploymentYaml) volumes() (interface{}, error) {
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
		Type     string       `json:"volume_type"`
		Name     string       `json:"volume_name"`
		Physical physicalInfo `json:"physical"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(dy.VolumeConf), &volumes); err != nil {
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

func (dy *DeploymentYaml) affinity() interface{} {
	/*
	  podAntiAffinity:
	    preferredDuringSchedulingIgnoredDuringExecution:
	    - podAffinityTerm:
	        labelSelector:
	          matchExpressions:
	          - key: service
	            operator: In
	            values:
	            - deployment名字
	        topologyKey: kubernetes.io/hostname
	      weight: 100
	*/

	// NOTE: 同一deployment下的pod散列在不同node上
	compare := map[string]interface{}{
		"key":      "service",
		"operator": "In",
		"values":   []string{dy.Deployment},
	}

	matchExpression := []interface{}{compare}
	labelSelector := map[string]interface{}{"matchExpressions": matchExpression}

	podAffinityTerm := map[string]interface{}{
		"labelSelector": labelSelector,
		"topologyKey":   "kubernetes.io/hostname",
	}

	policy := map[string]interface{}{
		"podAffinityTerm": podAffinityTerm,
		"weight":          100,
	}
	policies := []interface{}{policy}
	softLimit := map[string]interface{}{
		"preferredDuringSchedulingIgnoredDuringExecution": policies,
	}
	return map[string]interface{}{
		"podAntiAffinity": softLimit,
	}
}

func (dy *DeploymentYaml) containers() (interface{}, error) {
	/*
	  containers:
	  - name:
	    image:
	    imagePullPolicy:
	    env:
	    envFrom:
	    resources:
	      ...
	    securityContext:
	      ...
	    lifecycle:
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
	container["name"] = dy.Service
	container["image"] = fmt.Sprintf("%s:%s", dy.ImageURL, dy.ImageTag)
	container["imagePullPolicy"] = "IfNotPresent"
	container["env"] = dy.envs()
	container["envFrom"] = dy.envFrom()
	container["resources"] = dy.setResource(dy.QuotaCpu, dy.QuotaMaxCpu, dy.QuotaMem, dy.QuotaMaxMem)
	container["lifecycle"] = dy.lifecycle()
	container["securityContext"] = dy.security()
	volumeMounts, err := dy.mountContainerVolume()
	if err != nil {
		return nil, err
	}
	container["volumeMounts"] = volumeMounts
	container["livenessProbe"] = dy.liveness()
	container["readinessProbe"] = dy.readiness()
	container["terminationMessagePath"] = DEPLOY_TERMINATE_MESSAGE_PATH
	container["terminationMessagePolicy"] = DEPLOY_TERMINATE_POLICY
	containerList = append(containerList, container)
	return containerList, nil
}

func (dy *DeploymentYaml) envs() interface{} {
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

	env := []map[string]interface{}{
		{"name": "NAMESPACE", "value": dy.Namespace},
		{"name": "SERVICE", "value": dy.Service},
		{"name": "STAGE", "value": dy.Phase},
		{"name": "PODIP", "valueFrom": ipVal},
		{"name": "POD_NAME", "valueFrom": nameVal},
	}
	return env
}

func (dy *DeploymentYaml) envFrom() interface{} {
	/*
	  envFrom:
	  - configMapRef:
	    name: xxx
	*/
	pair := map[string]string{
		"name": dy.ConfigMap,
	}

	ref := map[string]interface{}{
		"configMapRef": pair,
	}

	froms := make([]map[string]interface{}, 0)
	froms = append(froms, ref)
	return froms
}

func (dy *DeploymentYaml) setResource(cpu, cpuMax, mem, memMax int) interface{} {
	/*
	  resources:
	   requests:
	     ...
	   limits:
	     ....
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

func (dy *DeploymentYaml) security() interface{} {
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

func (dy *DeploymentYaml) lifecycle() interface{} {
	/*
	   在容器被终结之前, Kubernetes 将发送一个 preStop 事件.
	   优雅关闭: 先发送一个kill信号(kill -3), 之后sleep 10秒等待未处理完的请求,
	             如果没处理完, 则会被强制终止(kill -9)
	   lifecycle:
	     preStop:
	       exec:
	         command:
	*/
	stopCmd := []string{
		"/bin/sh",
		"-c",
		"sleep 10",
	}
	stopExec := map[string]interface{}{"command": stopCmd}
	preStop := map[string]interface{}{"exec": stopExec}
	life := map[string]interface{}{"preStop": preStop}
	return life
}

func (dy *DeploymentYaml) mountContainerVolume() (interface{}, error) {
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
	if err := json.Unmarshal([]byte(dy.VolumeConf), &volumes); err != nil {
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

func (dy *DeploymentYaml) liveness() interface{} {
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

	return map[string]interface{}{
		"exec":                exec,
		"initialDelaySeconds": 5,
		"timeoutSeconds":      5,
		"periodSeconds":       60,
		"successThreshold":    1,
		"failureThreshold":    3,
	}
}

func (dy *DeploymentYaml) readiness() interface{} {
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

	return map[string]interface{}{
		"exec":                exec,
		"initialDelaySeconds": 5,
		"timeoutSeconds":      10,
		"periodSeconds":       10,
		"successThreshold":    1,
		"failureThreshold":    10,
	}
}
