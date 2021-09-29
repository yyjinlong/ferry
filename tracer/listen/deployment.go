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
	r := readiness(data, "Add")
	log.Infof("operate: Add deployment: %s readiness check %t", deployment, r)
}

func deploymentUpdate(oldObj, newObj interface{}) {
	newData := newObj.(*appsv1.Deployment)
	deployment := newData.ObjectMeta.Name
	r := readiness(newData, "Update")
	log.Infof("operate: Update deployment: %s readiness check %t", deployment, r)
	// TODO: 添加字段标识deployment是否发布完成.
	//       完成，进行蓝绿切换、endpoint记录
}

func deploymentDelete(obj interface{}) {
	data := obj.(*appsv1.Deployment)
	deployment := data.ObjectMeta.Name
	r := readiness(data, "Delete")
	log.Infof("operate: Delete deployment: %s readiness check %t", deployment, r)
}

func readiness(data *appsv1.Deployment, mode string) bool {
	deployment := data.ObjectMeta.Name

	metaData := data.ObjectMeta
	specData := data.Spec
	statData := data.Status

	metaGen := metaData.Generation
	statGen := statData.ObservedGeneration
	log.Infof("deployment: %s %s check generation meta: %d stat: %d", deployment, mode, metaGen, statGen)

	replicas := *specData.Replicas
	sReplicas := statData.Replicas
	uReplicas := statData.UpdatedReplicas
	aReplicas := statData.AvailableReplicas
	log.Infof("deployment: %s %s check replicas: %d", deployment, mode, replicas)
	log.Infof("deployment: %s %s check status_replicas: %d", deployment, mode, sReplicas)
	log.Infof("deployment: %s %s check updated_replicas: %d", mode, deployment, uReplicas)
	log.Infof("deployment: %s %s check avaliable_replicas: %d", mode, deployment, aReplicas)

	if statData.ObservedGeneration == metaData.Generation &&
		replicas == sReplicas &&
		replicas == uReplicas &&
		replicas == aReplicas {
		return true
	}
	return false
}
