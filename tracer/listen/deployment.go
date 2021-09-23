// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	appsv1 "k8s.io/api/apps/v1"

	"ferry/ops/log"
)

func deploymentAdd(obj interface{}) {
	data := obj.(*appsv1.Deployment)
	deployment := data.ObjectMeta.Name
	r := healthy(data, "Add")
	log.Infof("operate: Add deployment: %s healthy check %t", deployment, r)
}

func deploymentUpdate(oldObj, newObj interface{}) {
	newData := newObj.(*appsv1.Deployment)
	deployment := newData.ObjectMeta.Name
	r := healthy(newData, "Update")
	log.Infof("operate: Update deployment: %s healthy check %t", deployment, r)
}

func deploymentDelete(obj interface{}) {
	data := obj.(*appsv1.Deployment)
	deployment := data.ObjectMeta.Name
	r := healthy(data, "Delete")
	log.Infof("operate: Delete deployment: %s healthy check %t", deployment, r)
}

func healthy(data *appsv1.Deployment, mode string) bool {
	deployment := data.ObjectMeta.Name

	metaData := data.ObjectMeta
	specData := data.Spec
	statData := data.Status

	metaGen := metaData.Generation
	statGen := statData.ObservedGeneration
	log.Infof("operate: %s deployment: %s check generation meta: %d status: %d", mode, deployment, metaGen, statGen)

	replicas := *specData.Replicas
	sReplicas := statData.Replicas
	uReplicas := statData.UpdatedReplicas
	aReplicas := statData.AvailableReplicas
	log.Infof("operate: %s deployment: %s check replicas: %d status_replicas: %d updated_replicas: %d avaliable_replicas: %d",
		mode, deployment, replicas, sReplicas, uReplicas, aReplicas)

	if statData.ObservedGeneration == metaData.Generation &&
		replicas == sReplicas &&
		replicas == uReplicas &&
		replicas == aReplicas {
		return true
	}
	return false
}
