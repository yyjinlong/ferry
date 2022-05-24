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

	template, err := dy.template()
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

func (dy *DeploymentYaml) template() (map[string]interface{}, error) {
	/*
		template:
		  metadata:
		    ...
		  spec:
		    ...
	*/
	tpl := make(map[string]interface{})
	tpl["metadata"] = dy.templateMetadata()

	spec, err := dy.templateSpec()
	if err != nil {
		return nil, err
	}
	tpl["spec"] = spec
	return tpl, nil
}

func (dy *DeploymentYaml) templateMetadata() interface{} {
	labels := make(map[string]interface{})
	labels["labels"] = dy.labels()
	return labels
}

func (dy *DeploymentYaml) templateSpec() (interface{}, error) {
	/*
		spec:
		  hostAliases:
			...
		  dnsConfig:
			...
		  dnsPolicy:
		  imagePullecrets:
		    ...
		  nodeSelector:
			...
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
	spec["hostAliases"] = dy.hostAliases()
	spec["dnsConfig"] = dy.dnsConfig()
	spec["dnsPolicy"] = "None"
	spec["imagePullecrets"] = dy.imagePullecrets()
	spec["terminationGracePeriodSeconds"] = dy.ReserveTime
	spec["nodeSelector"] = dy.nodeSelector()

	volumes, err := dy.volumes()
	if err != nil {
		return nil, err
	}
	spec["volumes"] = volumes

	containers, err := dy.containers()
	if err != nil {
		return nil, err
	}
	spec["containers"] = containers
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
		hostAliases := make(map[string]interface{})
		hostAliases["ip"] = ip
		hostAliases["hostnames"] = hostList
		hostAliaseList = append(hostAliaseList, hostAliases)
	}
	return hostAliaseList
}

func (dy *DeploymentYaml) dnsConfig() interface{} {
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

func (dy *DeploymentYaml) imagePullecrets() interface{} {
	/*
		imagePullecrets:
		- name: xxx
	*/
	secrets := make([]map[string]string, 0)
	kv := map[string]string{
		"name": "harborkey",
	}
	secrets = append(secrets, kv)
	return secrets
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
	// NOTE: 在宿主机上创建本地存储卷, 目前只支持hostPath类型.
	volumes := make([]interface{}, 0)
	defineVolume, err := dy.createDefineVolume()
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, defineVolume)
	return volumes, nil
}

func (dy *DeploymentYaml) createDefineVolume() (interface{}, error) {
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
	if err := json.Unmarshal([]byte(dy.VolumeConf), &volumes); err != nil {
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

func (dy *DeploymentYaml) affinity() interface{} {
	/*
		同一deployment下的pod散列在不同node上.
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

	compare := make(map[string]interface{})
	compare["key"] = "service"
	compare["operator"] = "In"
	compare["values"] = []string{dy.Deployment}
	matchExpression := []interface{}{compare}
	labelSelector := map[string]interface{}{"matchExpressions": matchExpression}

	podAffinityTerm := make(map[string]interface{})
	podAffinityTerm["labelSelector"] = labelSelector
	podAffinityTerm["topologyKey"] = "kubernetes.io/hostname"

	policy := make(map[string]interface{})
	policy["podAffinityTerm"] = podAffinityTerm
	policy["weight"] = 100
	policies := []interface{}{policy}
	softLimit := map[string]interface{}{"preferredDuringSchedulingIgnoredDuringExecution": policies}
	return map[string]interface{}{"podAntiAffinity": softLimit}
}

func (dy *DeploymentYaml) containers() (interface{}, error) {
	/*
		- name:
		  image:
		  imagePullPolicy:
		  env:
		  envFrom:
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
	container["name"] = dy.Service
	container["image"] = fmt.Sprintf("%s:%s", dy.ImageURL, dy.ImageTag)
	container["imagePullPolicy"] = "IfNotPresent"
	container["env"] = dy.setEnv()
	container["envFrom"] = dy.envFrom()
	container["resources"] = dy.setResource(dy.QuotaCpu, dy.QuotaMaxCpu, dy.QuotaMem, dy.QuotaMaxMem)
	container["securityContext"] = dy.security()
	container["lifecycle"] = dy.lifecycle()
	volumeMounts, err := dy.mountContainerVolume()
	if err != nil {
		return nil, err
	}
	container["volumeMounts"] = volumeMounts
	container["livenessProbe"] = dy.liveness()
	container["readinessProbe"] = dy.readiness()
	container["terminationMessagePath"] = "/dev/termination-log"
	container["terminationMessagePolicy"] = "File"
	containerList = append(containerList, container)
	return containerList, nil
}

func (dy *DeploymentYaml) setEnv() interface{} {
	/*
	   - env:
	       - name:
	         value:
	*/
	env := []map[string]string{
		{"name": "SERVICE", "value": dy.Service},
	}
	return env
}

func (dy *DeploymentYaml) setResource(cpu, cpuMax, mem, memMax int) interface{} {
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

func (dy *DeploymentYaml) security() interface{} {
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
		Type      string        `json:"volume_type"`
		Name      string        `json:"volume_name"`
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

func (dy *DeploymentYaml) envFrom() interface{} {
	/*
		- configMapRef:
		   name: xxx
	*/
	froms := make([]map[string]interface{}, 0)

	pair := map[string]string{
		"name": dy.ConfigMap,
	}

	ref := map[string]interface{}{
		"configMapRef": pair,
	}
	froms = append(froms, ref)
	return froms
}
