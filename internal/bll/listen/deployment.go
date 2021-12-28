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

var (
	depResourceVersionMap = make(map[string]string)
)

func handleDeployment(obj interface{}, mode string) {
	data := obj.(*appsv1.Deployment)
	dep := &deployment{
		mode:              mode,
		deployment:        data.ObjectMeta.Name,
		resourceVersion:   data.ObjectMeta.ResourceVersion,
		replicas:          *data.Spec.Replicas,
		updatedReplicas:   data.Status.UpdatedReplicas,
		availableReplicas: data.Status.AvailableReplicas,
		metaGeneration:    data.ObjectMeta.Generation,
		statGeneration:    data.Status.ObservedGeneration,
	}
	dep.worker()
}

type deployment struct {
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

func (d *deployment) worker() {
	if !d.isValidDeployment() {
		return
	}
	if !d.isOldDeployment() {
		return
	}
	if !d.isReadiness() {
		return
	}

	log.Infof("[%s] deployment: %s is ready, replicas: %d", d.mode, d.deployment, d.replicas)
	if !d.parseServiceID() {
		return
	}
	if !d.parsePublishGroup() {
		return
	}

	key := fmt.Sprintf("%s_%s_%s", d.serviceName, d.phase, d.group)
	version, ok := depResourceVersionMap[key]
	if ok && version == d.resourceVersion {
		log.Infof("[%s] deployment: %s key: %s resource version: %s same so stop",
			d.mode, d.deployment, key, d.resourceVersion)
		return
	}
	depResourceVersionMap[key] = d.resourceVersion

	pipeline, err := objects.GetServicePipeline(d.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] deployment: %s query pipeline by service error: %s", d.mode, d.deployment, err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{db.PLSuccess, db.PLRollbackSuccess}) {
		delete(depResourceVersionMap, key) // NOTE: 上线完成删除对应的key
		log.Infof("[%s] deployment: %s deploy finish so stop", d.mode, d.deployment)
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
		oldDeployment = objects.GetDeployment(d.serviceName, pipeline.Service.ID, d.phase, offlineGroup)
	)

	switch d.mode {
	case Update:
		// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment缩成0
		if d.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, d.phase) {
			if err := d.scale(0, namespace, oldDeployment); err != nil {
				return
			}
			log.Infof("scale clear offline deployment: %s replicas: 0 success", oldDeployment)
		}
	}
	d.updateDBPhase(pipelineID)
}

func (d *deployment) isValidDeployment() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(d.deployment, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (d *deployment) isOldDeployment() bool {
	if d.replicas == 0 {
		return false
	}
	return true
}

func (d *deployment) isReadiness() bool {
	if d.metaGeneration == d.statGeneration &&
		d.replicas == d.updatedReplicas && d.replicas == d.availableReplicas {
		return true
	}
	return false
}

func (d *deployment) parseServiceID() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(d.deployment, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service id convert to int64 error: %s", err)
		return false
	}
	d.serviceID = serviceID
	log.Infof("[%s] deployment: %s get service id: %d", d.mode, d.deployment, d.serviceID)
	return true
}

func (d *deployment) parsePublishGroup() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}
	matchList := reg.Split(d.deployment, -1)
	afterList := strings.Split(matchList[1], "-")
	d.serviceName = matchList[0]
	d.phase = afterList[0]
	d.group = afterList[1]
	log.Infof("[%s] deployment: %s service name: %s phase: %s group: %s",
		d.mode, d.deployment, d.serviceName, d.phase, d.group)
	return true
}

func (d *deployment) scale(replicas int, namespace, deployment string) error {
	var (
		url     = fmt.Sprintf(g.Config().K8S.Deployment, namespace) + "/" + deployment + "/scale"
		header  = map[string]string{"Content-Type": "application/strategic-merge-patch+json"}
		payload = fmt.Sprintf(`{"spec": {"replicas": %d}}`, replicas)
	)

	body, err := g.Patch(url, header, []byte(payload), 5)
	if err != nil {
		log.Errorf("[%s] scale deployment: %s replicas: %d error: %s", d.mode, deployment, replicas, err)
		return err
	}

	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("[%s] scale deployment: %s response json decode error: %s", d.mode, deployment, err)
		return err
	}

	spec := resp["spec"].(map[string]interface{})
	if len(spec) != 0 && spec["replicas"].(float64) == float64(replicas) {
		log.Infof("[%s] scale deployment: %s replicas: %d success.", d.mode, deployment, replicas)
	}
	return nil
}

func (d *deployment) updateDBPhase(pipelineID int64) {
	err := objects.UpdatePhase(pipelineID, d.phase, db.PHSuccess)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] deployment: %s update pipeline: %d phase: %s failed: %s",
			d.mode, d.deployment, pipelineID, d.phase, err)
	}
}
