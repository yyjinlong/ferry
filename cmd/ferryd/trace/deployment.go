// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package trace

import (
	"encoding/json"
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
	serviceID         int64
	serviceName       string
	phase             string
	group             string
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
	de.parse()

	pipeline, err := objects.GetServicePipeline(de.serviceID)
	if err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return
	}

	if g.Ini(pipeline.Pipeline.Status, []int{db.PLSuccess, db.PLRollbackSuccess}) {
		log.Infof("query pipeline: %d deploy finished, so ignore.", pipeline.Pipeline.ID)
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

	if de.mode == "update" {
		// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
		if de.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, de.phase) {
			log.Infof("scale clear offline group is: %s", offlineGroup)
			log.Infof("scale clear offline deployment is: %s", oldDeployment)
			if err := de.scale(0, namespace, oldDeployment); err != nil {
				return
			}
			log.Infof("scale clear offline deployment: %s replicas: 0 success", oldDeployment)
		}
	}
	de.updateDB(pipelineID)
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

func (de *DeploymentEvent) parse() {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return
	}

	result := reg.FindAllStringSubmatch(de.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service id convert to int64 error: %s", err)
		return
	}
	de.serviceID = serviceID

	matchList := reg.Split(de.deployment, -1)
	afterList := strings.Split(matchList[1], "-")
	de.serviceName = matchList[0]
	de.phase = afterList[0]
	de.group = afterList[1]
	log.Infof("get service name: %s phase: %s group: %s", de.serviceName, de.phase, de.group)
}

func (de *DeploymentEvent) updateDB(pipelineID int64) {
	if err := objects.UpdatePhase(pipelineID, de.phase, db.PHSuccess); err != nil {
		log.Errorf("update pipeline: %d phase: %s failed: %s", pipelineID, de.phase, err)
		return
	}
	log.Infof("update pipeline: %d phase: %s success", pipelineID, de.phase)
}

func (de *DeploymentEvent) scale(replicas int, namespace, deployment string) error {
	var (
		url     = fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment + "/scale"
		header  = map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
		payload = fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	)

	body, err := g.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("scale deployment: %s replicas: %d error: %s", deployment, replicas, err)
		return err
	}

	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("scale deployment: %s response json decode error: %s", deployment, err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 && spec["replicas"].(float64) == float64(replicas) {
		log.Infof("scale deployment: %s replicas: %d success.", deployment, replicas)
	}
	return nil
}
