// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"errors"
	"fmt"
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
	fe := &finishEvent{
		mode:              mode,
		deployment:        data.ObjectMeta.Name,
		resourceVersion:   data.ObjectMeta.ResourceVersion,
		replicas:          *data.Spec.Replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	}
	fe.worker()
}

type finishEvent struct {
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

func (fe *finishEvent) worker() {
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "mode": fe.mode, "deployment": fe.deployment, "version": fe.resourceVersion})
	if !fe.isValidDeployment() {
		return
	}
	if !fe.isOldDeployment() {
		return
	}
	if !fe.isReadiness() {
		return
	}

	log.Infof("deployment is ready, replicas: %d", fe.replicas)
	if !fe.parseServiceID() {
		return
	}
	if !fe.parsePublishGroup() {
		return
	}

	key := fmt.Sprintf("%s_%s_%s", fe.serviceName, fe.phase, fe.group)
	version, ok := depResourceVersionMap[key]
	if ok && version == fe.resourceVersion {
		log.Infof("check key: %s resource version: %s same so stop", key, fe.resourceVersion)
		return
	}
	depResourceVersionMap[key] = fe.resourceVersion
	log.Infof("get service: %s phase: %s group: %s", fe.serviceName, fe.phase, fe.group)

	pipeline, err := objects.GetServicePipeline(fe.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		delete(depResourceVersionMap, key) // NOTE: 上线完成删除对应的key
		log.Info("check deploy is finished so stop")
		return
	}
	fe.execute(pipeline)
}

func (fe *finishEvent) isValidDeployment() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(fe.deployment, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (fe *finishEvent) isOldDeployment() bool {
	if fe.replicas == 0 {
		return false
	}
	return true
}

func (fe *finishEvent) isReadiness() bool {
	if fe.metaGeneration == fe.statGeneration &&
		fe.replicas == fe.updatedReplicas && fe.replicas == fe.availableReplicas {
		return true
	}
	return false
}

func (fe *finishEvent) parseServiceID() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(fe.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service id convert to int64 error: %s", err)
		return false
	}
	fe.serviceID = serviceID
	return true
}

func (fe *finishEvent) parsePublishGroup() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	matchList := reg.Split(fe.deployment, -1)
	afterList := strings.Split(matchList[1], "-")
	fe.serviceName = matchList[0]
	fe.phase = afterList[0]
	fe.group = afterList[1]
	return true
}

func (fe *finishEvent) execute(pipeline *model.PipelineQuery) {
	var (
		pipelineID   = pipeline.Pipeline.ID
		namespace    = pipeline.Namespace.Name
		offlineGroup = pipeline.Service.OnlineGroup // NOTE: 在确认时, 原有表记录的组则变为待下线组
	)

	if offlineGroup == "" {
		return
	}

	var (
		publishGroup  = objects.GetDeployGroup(offlineGroup)
		oldDeployment = objects.GetDeployment(fe.serviceName, pipeline.Service.ID, fe.phase, offlineGroup)
	)

	// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
	if fe.mode == Update && fe.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, fe.phase) {
		dep := k8s.NewDeployments(namespace, oldDeployment)
		if err := dep.Scale(0); err != nil {
			return
		}
		log.Infof("phase: %s scale offline deployment: %s replicas: 0 success", fe.phase, oldDeployment)
	}

	if err := objects.UpdatePhase(pipelineID, fe.phase, model.PHSuccess); !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("phase: %s update pipeline: %d  failed: %s", fe.phase, pipelineID, err)
		return
	}
	log.Infof("phase: %s update pipeline: %d to db success.", fe.phase, pipelineID)
}
