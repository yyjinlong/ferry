// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/cfg"
	"nautilus/pkg/cm"
	"nautilus/pkg/k8s/exec"
	"nautilus/pkg/k8s/yaml"
	"nautilus/pkg/model"
)

func NewDeploy() *Deploy {
	return &Deploy{}
}

type Deploy struct{}

func (d *Deploy) Handle(pid int64, phase, username string) error {
	// TODO: 建立websocket

	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(cfg.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	if cm.Ini(pipeline.Status, []int{model.PLSuccess, model.PLFailed, model.PLRollbackSuccess, model.PLRollbackFailed, model.PLTerminate}) {
		return fmt.Errorf(cfg.PUB_DEPLOY_FINISHED)
	}

	serviceID := pipeline.ServiceID
	svcObj, err := model.GetServiceByID(serviceID)
	if err != nil {
		return fmt.Errorf(cfg.DB_SERVICE_QUERY_ERROR, err)
	}
	serviceName := svcObj.Name
	deployGroup := svcObj.DeployGroup

	nsObj, err := model.GetNamespace(svcObj.NamespaceID)
	if err != nil {
		return fmt.Errorf(cfg.DB_QUERY_NAMESPACE_ERROR, err)
	}
	namespace := nsObj.Name

	var (
		deployment = cm.GetDeployment(serviceName, serviceID, phase, deployGroup)
		appid      = cm.GetAppID(serviceName, serviceID, phase)
	)
	log.Infof("current deploy group: %s deployment name: %s appid: %s", deployGroup, deployment, appid)

	imageInfo, err := model.FindImageInfo(pid)
	if err != nil {
		return fmt.Errorf(cfg.DB_PIPELINE_UPDATE_ERROR, err)
	}

	if len(imageInfo) == 0 {
		return fmt.Errorf("get image info is empty")
	}
	log.Infof("create yaml get image info: %s", imageInfo)

	replicas := svcObj.Replicas
	if phase == model.PHASE_SANDBOX {
		// NOTE: 沙盒阶段默认返回1个副本
		replicas = 1
	}

	// 创建yaml
	depYaml := &yaml.DeploymentYaml{
		Phase:       phase,
		Deployment:  deployment,
		AppID:       appid,
		Namespace:   namespace,
		Service:     serviceName,
		ImageURL:    imageInfo["image_url"],
		ImageTag:    imageInfo["image_tag"],
		Replicas:    replicas,
		QuotaCpu:    svcObj.QuotaCpu,
		QuotaMaxCpu: svcObj.QuotaMaxCpu,
		QuotaMem:    svcObj.QuotaMem,
		QuotaMaxMem: svcObj.QuotaMaxMem,
		VolumeConf:  svcObj.Volume,
		ReserveTime: svcObj.ReserveTime,
	}
	tpl, err := depYaml.Instance()
	if err != nil {
		return fmt.Errorf(cfg.PUB_BUILD_DEPLOYMENT_YAML_ERROR, err)
	}
	log.Infof("generate deployment yaml(%s) success", deployment)
	fmt.Println(tpl)

	if err := d.execute(namespace, deployment, tpl); err != nil {
		return fmt.Errorf(cfg.PUB_K8S_DEPLOYMENT_EXEC_FAILED, err)
	}
	log.Infof("pubish deployment: %s to k8s success", deployment)

	if err := model.CreatePhase(pid, model.PHASE_DEPLOY, phase, model.PHProcess, tpl); err != nil {
		return fmt.Errorf(cfg.PUB_RECORD_DEPLOYMENT_TO_DB_ERROR, err)
	}
	log.Infof("record deployment: %s to db success", deployment)
	return nil
}

func (d *Deploy) execute(namespace, deployment, tpl string) error {
	dep := exec.NewDeployments(namespace, deployment)
	if !dep.Exist() {
		return dep.Create(tpl)
	}
	return dep.Update(tpl)
}
