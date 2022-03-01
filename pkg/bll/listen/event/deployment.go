// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"

	"nautilus/internal/k8s"
	"nautilus/internal/model"
	"nautilus/internal/objects"
	"nautilus/pkg/g"
	"nautilus/pkg/log"
)

func HandleDeploymentCapturer(obj interface{}, mode string) {
	data := obj.(*appsv1.Deployment)
	var (
		deployment      = data.ObjectMeta.Name
		resourceVersion = data.ObjectMeta.ResourceVersion
		replicas        = *data.Spec.Replicas
	)

	log.InitFields(log.Fields{
		"logid":      g.UniqueID(),
		"mode":       mode,
		"deployment": deployment,
		"version":    resourceVersion,
	})

	handleEvent(&deploymentCapturer{
		mode:              mode,
		deployment:        deployment,
		resourceVersion:   resourceVersion,
		replicas:          replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	})
}

type deploymentCapturer struct {
	mode              string
	deployment        string
	resourceVersion   string
	replicas          int32
	updatedReplicas   int32
	availableReplicas int32
	metaGeneration    int64
	statGeneration    int64
	serviceID         int64
	serviceName       string
	phase             string
	group             string
}

func (c *deploymentCapturer) valid() bool {
	// NOTE: 检测是否是业务的deployment
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(c.deployment, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (c *deploymentCapturer) ready() bool {
	// NOTE: 忽略副本数为0的deployment
	if c.replicas == 0 {
		return false
	}

	// NOTE: 检查deployment是否就绪(spec中的副本数是否等于status中的副本数)
	if !(c.metaGeneration == c.statGeneration &&
		c.replicas == c.updatedReplicas && c.replicas == c.availableReplicas) {
		return false
	}
	log.Infof("check deployment is ready, replicas: %d", c.replicas)
	return true
}

func (c *deploymentCapturer) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}

	// 获取服务名、阶段、蓝绿组
	matchList := reg.Split(c.deployment, -1)
	c.serviceName = matchList[0]

	afterList := strings.Split(matchList[1], "-")
	c.phase = afterList[0]
	c.group = afterList[1]

	// 获取服务ID
	result := reg.FindAllStringSubmatch(c.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service id convert to int64 error: %s", err)
		return false
	}
	c.serviceID = serviceID
	return true
}

func (c *deploymentCapturer) operate() bool {
	pipeline, err := objects.GetServicePipeline(c.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return false
	}

	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Info("check deploy is finished so stop")
		return false
	}

	var (
		pipelineID = pipeline.Pipeline.ID
		namespace  = pipeline.Namespace.Name
	)

	kind := model.PHASE_DEPLOY
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLRollbacking, model.PLRollbackSuccess, model.PLRollbackFailed}) {
		kind = model.PHASE_ROLLBACK
	}

	log.Infof("get pipeline: %d kind: %s phase: %s", pipelineID, kind, c.phase)
	phaseObj, err := objects.GetPhaseInfo(pipelineID, kind, c.phase)
	if errors.Is(err, objects.NotFound) {
		return false
	}
	if err != nil {
		log.Errorf("query phase info error: %s", err)
		return false
	}

	// 判断resourceVersion是否相同
	if phaseObj.ResourceVersion == c.resourceVersion {
		return true
	}

	// 判断该阶段是否完成
	if g.Ini(phaseObj.Status, []int{model.PHSuccess, model.PHFailed}) {
		return true
	}

	// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
	if c.mode == Update && c.group == pipeline.Service.DeployGroup && objects.CheckPhaseIsDeploy(pipelineID, kind, c.phase) {
		oldDeployment := objects.GetDeployment(c.serviceName, c.serviceID, c.phase, pipeline.Service.OnlineGroup)
		dep := k8s.NewDeployments(namespace, oldDeployment)
		if err := dep.Scale(0); err != nil {
			return false
		}
		log.Infof("---- scale deployment: %s replicas: 0 on phase: %s success", oldDeployment, c.phase)
	}

	if err := objects.UpdatePhaseV2(pipelineID, kind, c.phase, model.PHSuccess, c.resourceVersion); err != nil {
		log.Errorf("update pipeline: %d to db on phase: %s failed: %s", pipelineID, c.phase, err)
		return false
	}
	log.Infof("update pipeline: %d to db on phase: %s success.", pipelineID, c.phase)
	return true
}
