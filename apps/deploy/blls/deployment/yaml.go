// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"ferry/ops/log"
)

type yaml struct {
	pipelineID    int64               // 流水线ID
	phase         string              // 部署阶段
	namespace     string              // 当前服务所在命名空间
	deployment    string              // deployment名字
	serviceName   string              // 服务名
	deployPath    string              // 部署路径
	replicas      int                 // 副本数
	reserveTime   int                 // 终止后的预留时间
	containerConf string              // 容器配置
	volumeConf    string              // 数据卷配置
	imageList     []map[string]string // 模块镜像信息列表
	rootPath      string              // 容器内默认部署根路径
	volumeName    string              // 容器内默认卷名称
}

func (y *yaml) init() {
	// NOTE: 服务部署路径默认为: hostPath类型的根路径.
	y.rootPath = y.deployPath
	log.Infof("default volume root path: %s", y.rootPath)

	// NOTE: 基于根路径最后一段设置为: hostPath类型的默认数据卷名称.
	y.volumeName = filepath.Base(y.rootPath)
	log.Infof("default volume name: %s", y.volumeName)
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
		"service": y.serviceName,
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
	*/
	spec := make(map[string]interface{})
	spec["hostAliases"] = y.hostAliases()
	spec["dnsConfig"] = y.dnsConfig()
	spec["dnsPolicy"] = "None"
	spec["terminationGracePeriodSeconds"] = y.reserveTime

	// NOTE: 加载默认的数据卷; 服务配置的数据卷.
	volumes, err := y.volumes()
	if err != nil {
		return nil, err
	}
	spec["volumes"] = volumes

	// NOTE: 初始化该服务的各个模块代码.
	// NOTE: pod里的存储卷是共享的, 所以initContainers产生的数据可以被主容器访问到
	spec["initContainers"] = y.initContainers()

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
	// 在宿主机上创建本地存储卷, 目前只支持hostPath类型.·
	volumes := make([]interface{}, 0)
	volumes = append(volumes, y.createDefaultVolume())

	defineVolume, err := y.createDefineVolume()
	if err != nil {
		return nil, err
	}
	volumes = append(volumes, defineVolume)
	return volumes, nil
}

// 创建默认数据卷
func (y *yaml) createDefaultVolume() interface{} {
	/*
		volumes:
		  - name: www
		    hostPath:
		      path: /home/worker/www/ivr
		      type: DirectoryOrCreate
		说明: 将宿主机上的/home/worker/www/ivr目录挂载到pod内, 挂载点名为www
		约定: 宿主机的根路径 == 容器内服务的根路径
	*/
	nodeRootPath := y.rootPath
	nodeHostPath := fmt.Sprintf("%s/%s/%d", nodeRootPath, y.serviceName, y.pipelineID)
	hostPath := map[string]string{
		"path": nodeHostPath,
		"type": "DirectoryOrCreate",
	}
	defaultVolume := make(map[string]interface{})
	defaultVolume["name"] = y.volumeName
	defaultVolume["hostPath"] = hostPath
	log.Infof("create default volume: %s finish", defaultVolume)
	return defaultVolume
}

// 创建自定义的数据卷(服务需要的数据卷)
func (y *yaml) createDefineVolume() (interface{}, error) {
	/*
	   volumes:
	     - name:
	       hostPath:
	         path:
	         type:
	*/

	type volume struct {
		VolumeName   string `json:"newvolume_name"`
		VolumeType   string `json:"newvolume_type"`
		HostPathType string `json:"hostpath_type"`
		HostPath     string `json:"hostpath"`
	}
	var volumeList []volume

	if err := json.Unmarshal([]byte(y.volumeConf), &volumeList); err != nil {
		return nil, err
	}

	defineVolume := make(map[string]interface{})
	for _, item := range volumeList {
		defineVolume["name"] = item.VolumeName
		if item.VolumeType == "hostPath" {
			hostPath := map[string]string{
				"type": item.HostPathType,
				"path": item.HostPath,
			}
			defineVolume["hostPath"] = hostPath
		}
	}
	log.Infof("create define volume: %s finish", defineVolume)
	return defineVolume, nil
}

func (y *yaml) initContainers() interface{} {
	/*
		initContainers:
		  - name:
		    image:
		    imagePullPolicy:
		    command:
		    ...
		    volumeMounts:
		    ...
	*/
	initContainers := make([]interface{}, 0)
	for _, item := range y.imageList {
		moduleName := item["module_name"]
		imageURL := item["image_url"]
		imageTag := item["image_tag"]
		log.Infof("init container current init module: %s", moduleName)

		// NOTE: 将容器内的部署路径挂载到默认的挂载点下
		volumeMounts := []map[string]string{{
			"name":      y.volumeName,
			"mountPath": y.rootPath,
		}}
		log.Infof("init container mount path: %s to: %s", y.rootPath, y.volumeName)

		// NOTE: 代码镜像添加'-code'后缀, 区别于运行镜像名, 避免重名.
		containerName := fmt.Sprintf("%s-code", moduleName)
		containerInfo := map[string]interface{}{
			"volumeMounts":    volumeMounts,
			"name":            containerName,
			"image":           fmt.Sprintf("%s:%s", imageURL, imageTag),
			"imagePullPolicy": "IfNotPresent",
			"command":         y.codeCopy(moduleName),
		}
		initContainers = append(initContainers, containerInfo)
	}
	return initContainers
}

// 将代码拷贝到数据卷所挂载的节点路径
func (y *yaml) codeCopy(moduleName string) []string {
	lockFile := fmt.Sprintf("%s/cp_code_lock_%s", y.rootPath, moduleName)
	doneFile := fmt.Sprintf("%s/cp_code_done_%s", y.rootPath, moduleName)
	destPath := filepath.Join(y.rootPath, moduleName)

	copyCmdList := []string{
		"if [ ! -f \"" + doneFile + "\" ]; then",
		"cp -rfp /src/* " + y.rootPath + " &&",
		"chown -R 500:500 " + destPath + " &&",
		"touch " + doneFile + "; fi",
	}
	copyCmd := strings.Join(copyCmdList, " ")

	command := []string{
		"/usr/bin/flock",
		"-x",
		lockFile,
		"-c",
		copyCmd,
	}
	return command
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
	*/
	type volumeInfo struct {
		Name      string `json:"volume_name"`
		MountPath string `json:"volume_mount_path"`
		SubPath   string `json:"volume_subpath"`
	}

	type quotaInfo struct {
		CPU    int `json:"budget_cpu"`
		CPUMax int `json:"budget_max_cpu"`
		Mem    int `json:"budget_memory"`
		MemMax int `json:"budget_max_memory"`
	}

	type containerInfo struct {
		ImageURL string       `json:"image_url"`
		Volume   []volumeInfo `json:"volume"`
		Quota    quotaInfo    `json:"quota"`
	}

	var cList []containerInfo

	if err := json.Unmarshal([]byte(y.containerConf), &cList); err != nil {
		return nil, err
	}

	containerList := make([]interface{}, 0)
	for _, item := range cList {
		resources := y.setResource(item.Quota.CPU, item.Quota.CPUMax, item.Quota.Mem, item.Quota.MemMax)

		containers := make(map[string]interface{})
		containers["volumeMounts"] = y.mountVolume()
		containers["name"] = y.serviceName
		containers["image"] = item.ImageURL
		containers["imagePullPolicy"] = "IfNotPresent"
		containers["env"] = y.setEnv()
		containers["resources"] = resources
		containers["securityContext"] = y.security()
		containers["lifecycle"] = y.lifecycle()

		// livenessProbe

		// readinessProbe

		containerList = append(containerList, containers)
	}
	return containerList, nil
}

func (y *yaml) setEnv() interface{} {
	/*
	   - env:
	       - name:
	         value:
	*/
	env := []map[string]string{
		{"name": "SERVICE", "value": y.serviceName},
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

// 将宿主机代码挂载到主容器
func (y *yaml) mountVolume() interface{} {
	/*
	   volumeMounts:
	     - mountPath:
	       name:
	       subPath:
	*/
	rootPath := y.rootPath     // 容器内根路径
	mountPoint := y.volumeName // 容器内挂载点

	containerVolumeMounts := make([]map[string]string, 0)
	for _, item := range y.imageList {
		moduleName := item["module_name"]
		mountPath := filepath.Join(rootPath, moduleName)
		containerVolume := map[string]string{
			"name":      mountPoint,
			"mountPath": mountPath,
			"subPath":   moduleName,
		}
		containerVolumeMounts = append(containerVolumeMounts, containerVolume)
		log.Infof("container volume mount path: %s to: %s", mountPath, mountPoint)
	}
	return containerVolumeMounts
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
	   lifecycle:
	     preStop:
	       exec:
	         command:
	*/
	stopCmd := []string{
		"/bin/sh",
		"-c",
		"sleep 2",
	}
	stopExec := map[string]interface{}{"command": stopCmd}
	preStop := map[string]interface{}{"exec": stopExec}

	life := make(map[string]interface{})
	life["preStop"] = preStop
	return life
}
