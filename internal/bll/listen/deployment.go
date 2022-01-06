// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"

	"ferry/internal/k8s"
	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

func CheckDeploymentIsFinish(obj interface{}, mode string) {
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

	dh := &dephandler{
		mode:              mode,
		deployment:        deployment,
		resourceVersion:   resourceVersion,
		replicas:          replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	}
	if !dh.valid() {
		return
	}
	if !dh.parse() {
		return
	}
	log.Infof("check deployment is ready, replicas: %d", dh.replicas)

	if !dh.operate() {
		return
	}
	log.Infof("check deployment is finished")
}

type dephandler struct {
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

func (h *dephandler) valid() bool {
	// NOTE: 检测是否是业务的deployment
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(h.deployment, -1)
	if len(result) == 0 {
		return false
	}

	// NOTE: 忽略副本数为0的deployment
	if h.replicas == 0 {
		return false
	}

	// NOTE: 检查deployment是否就绪
	if !(h.metaGeneration == h.statGeneration && h.replicas == h.updatedReplicas && h.replicas == h.availableReplicas) {
		return false
	}
	return true
}

func (h *dephandler) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}

	// 获取服务名、阶段、蓝绿组
	matchList := reg.Split(h.deployment, -1)
	h.serviceName = matchList[0]

	afterList := strings.Split(matchList[1], "-")
	h.phase = afterList[0]
	h.group = afterList[1]

	// 获取服务ID
	result := reg.FindAllStringSubmatch(h.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service id convert to int64 error: %s", err)
		return false
	}
	h.serviceID = serviceID
	return true
}

func (h *dephandler) operate() bool {
	pipeline, err := objects.GetServicePipeline(h.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return false
	}

	// 判断该上线流程是否完成
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Info("check deploy is finished so stop")
		return false
	}

	var (
		pipelineID   = pipeline.Pipeline.ID
		namespace    = pipeline.Namespace.Name
		offlineGroup = pipeline.Service.OnlineGroup // NOTE: 当前在线的组为待下线组
	)

	if offlineGroup == "" {
		return true
	}

	kind := model.PHASE_DEPLOY
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLRollbacking, model.PLRollbackSuccess, model.PLRollbackFailed}) {
		kind = model.PHASE_ROLLBACK
	}

	log.Infof("get pipeline: %d kind: %s phase: %s", pipelineID, kind, h.phase)
	phaseObj, err := objects.GetPhaseInfo(pipelineID, model.PHASE_DEPLOY, h.phase)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query phase info error: %s", err)
		return false
	}

	// 判断resourceVersion是否相同
	if phaseObj.ResourceVersion == h.resourceVersion {
		return true
	}

	// 判断该阶段是否完成
	if g.Ini(phaseObj.Status, []int{model.PHSuccess, model.PHFailed}) {
		return true
	}

	var (
		publishGroup  = objects.GetDeployGroup(offlineGroup)
		oldDeployment = objects.GetDeployment(h.serviceName, h.serviceID, h.phase, offlineGroup)
	)

	// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
	if h.mode == Update && h.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, h.phase) {
		dep := k8s.NewDeployments(namespace, oldDeployment)
		if err := dep.Scale(0); err != nil {
			return false
		}
		log.Infof("---- scale deployment: %s replicas: 0 on phase: %s success", oldDeployment, h.phase)
	}

	if err := objects.UpdatePhaseV2(pipelineID, kind, h.phase, model.PHSuccess, h.resourceVersion); err != nil {
		log.Errorf("update pipeline: %d to db on phase: %s failed: %s", pipelineID, h.phase, err)
		return false
	}
	log.Infof("update pipeline: %d to db on phase: %s success.", pipelineID, h.phase)
	return true
}
