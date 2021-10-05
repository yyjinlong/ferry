// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"regexp"

	appsv1 "k8s.io/api/apps/v1"

	"ferry/ops/log"
	"ferry/tracer/handle"
)

func DeploymentAdd(obj interface{}) {
	depEvent := NewDeploymentEvent(obj, "add")
	depEvent.Run()
}

func DeploymentUpdate(oldObj, newObj interface{}) {
	depEvent := NewDeploymentEvent(newObj, "update")
	depEvent.Run()
}

func DeploymentDelete(obj interface{}) {
	// NOTE: deployment删除暂不处理
}

func NewDeploymentEvent(obj interface{}, mode string) *DeploymentEvent {
	data := obj.(*appsv1.Deployment)
	return &DeploymentEvent{
		mode:              mode,
		deployment:        data.ObjectMeta.Name,
		replicas:          *data.Spec.Replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	}
}

type DeploymentEvent struct {
	mode              string
	deployment        string
	replicas          int32
	updatedReplicas   int32
	availableReplicas int32
	metaGeneration    int64
	statGeneration    int64
}

func (de *DeploymentEvent) Run() {
	if !de.isValidDeployment() {
		return
	}
	if !de.isOldDeployment() {
		return
	}
	if !de.isReadiness() {
		return
	}
	de.worker()
}

func (de *DeploymentEvent) isValidDeployment() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(de.deployment, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (de *DeploymentEvent) isOldDeployment() bool {
	if de.replicas == 0 {
		log.Infof("%s deployment: %s replicas is 0, ignore!", de.mode, de.deployment)
		return false
	}
	return true
}

func (de *DeploymentEvent) isReadiness() bool {
	if de.metaGeneration == de.statGeneration &&
		de.replicas == de.updatedReplicas && de.replicas == de.availableReplicas {
		log.Infof("%s deployment: %s is readiness. generation: %d replicas: %d",
			de.mode, de.deployment, de.metaGeneration, de.replicas)
		return true
	}
	log.Infof("%s deployment: %s is not ready", de.mode, de.deployment)
	return false
}

func (de *DeploymentEvent) worker() {
	deploymentHandler := handle.NewHandleDeployment(de.deployment)
	if de.mode == "update" {
		deploymentHandler.ClearOld()
	}
}
