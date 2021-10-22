// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package trace

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"

	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/server/db"
	"ferry/server/objects"
)

func handleDeployment(obj interface{}, mode string) {
	data := obj.(*appsv1.Deployment)
	dep := &deploymentEvent{
		mode:              mode,
		deployment:        data.ObjectMeta.Name,
		replicas:          *data.Spec.Replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	}
	dep.worker()
}

type deploymentEvent struct {
	mode              string
	deployment        string
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

func (de *deploymentEvent) worker() {
	if !de.isValidDeployment() {
		return
	}
	if !de.isOldDeployment() {
		return
	}
	if !de.isReadiness() {
		return
	}
	log.Infof("[%s] deployment: %s is ready, replicas: %d", de.mode, de.deployment, de.replicas)

	if !de.getServiceID() {
		return
	}
	if !de.getPublishGroup() {
		return
	}

	pipeline, err := objects.GetServicePipeline(de.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] query pipeline by service error: %s", de.mode, err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{db.PLSuccess, db.PLRollbackSuccess}) {
		return
	}

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
		oldDeployment = objects.GetDeployment(de.serviceName, pipeline.Service.ID, de.phase, offlineGroup)
	)

	switch de.mode {
	case Update:
		// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
		if de.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, de.phase) {
			if err := de.scale(0, namespace, oldDeployment); err != nil {
				return
			}
			log.Infof("scale clear offline deployment: %s replicas: 0 success", oldDeployment)
		}
		fallthrough
	default:
		de.updateDB(pipelineID)
	}
}

func (de *deploymentEvent) isValidDeployment() bool {
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

func (de *deploymentEvent) isOldDeployment() bool {
	if de.replicas == 0 {
		return false
	}
	return true
}

func (de *deploymentEvent) isReadiness() bool {
	if de.metaGeneration == de.statGeneration &&
		de.replicas == de.updatedReplicas && de.replicas == de.availableReplicas {
		return true
	}
	return false
}

func (de *deploymentEvent) getServiceID() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(de.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service id convert to int64 error: %s", err)
		return false
	}
	de.serviceID = serviceID
	log.Infof("[%s] deployment: %s get service id: %d", de.mode, de.deployment, de.serviceID)
	return true
}

func (de *deploymentEvent) getPublishGroup() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	matchList := reg.Split(de.deployment, -1)
	afterList := strings.Split(matchList[1], "-")
	de.serviceName = matchList[0]
	de.phase = afterList[0]
	de.group = afterList[1]
	log.Infof("[%s] deployment: %s service name: %s phase: %s group: %s",
		de.mode, de.deployment, de.serviceName, de.phase, de.group)
	return true
}

func (de *deploymentEvent) scale(replicas int, namespace, deployment string) error {
	var (
		url     = fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment + "/scale"
		header  = map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
		payload = fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	)

	body, err := g.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("[%s] scale deployment: %s replicas: %d error: %s", de.mode, deployment, replicas, err)
		return err
	}

	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("[%s] scale deployment: %s response json decode error: %s", de.mode, deployment, err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 && spec["replicas"].(float64) == float64(replicas) {
		log.Infof("[%s] scale deployment: %s replicas: %d success.", de.mode, deployment, replicas)
	}
	return nil
}

func (de *deploymentEvent) updateDB(pipelineID int64) {
	err := objects.UpdatePhase(pipelineID, de.phase, db.PHSuccess)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("update pipeline: %d phase: %s failed: %s", pipelineID, de.phase, err)
		return
	}
}
