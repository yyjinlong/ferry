// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package handle

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/objects"
)

func NewHandleDeployment(deployment string) *HandleDeployment {
	return &HandleDeployment{
		deployment: deployment,
	}
}

type HandleDeployment struct {
	deployment  string
	serviceID   int64
	serviceName string
	phase       string
	group       string
}

func (hd *HandleDeployment) ClearOld() {
	hd.parse()

	pipeline, err := objects.GetServicePipeline(hd.serviceID)
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
		oldDeployment = objects.GetDeployment(hd.serviceName, pipeline.Service.ID, hd.phase, offlineGroup)
	)

	// 如果就绪的是当前的部署组, 并且对应该阶段也正在发布, 则需要将旧的deployment清零
	if hd.group == publishGroup && objects.CheckPhaseIsDeploy(pipelineID, hd.phase) {
		log.Infof("scale clear offline group is: %s", offlineGroup)
		log.Infof("scale clear offline deployment is: %s", oldDeployment)
		if err := hd.scale(0, namespace, oldDeployment); err != nil {
			return
		}
		log.Infof("scale clear offline deployment: %s replicas: 0 success", oldDeployment)

		if err := objects.UpdatePhase(pipelineID, hd.phase, db.PHSuccess); err != nil {
			log.Errorf("update pipeline: %d phase: %s failed: %s", pipelineID, hd.phase, err)
			return
		}
		log.Infof("update pipeline: %d phase: %s success", pipelineID, hd.phase)
	}
}

func (hd *HandleDeployment) parse() {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return
	}

	result := reg.FindAllStringSubmatch(hd.deployment, -1)
	if len(result) == 0 {
		return
	}

	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service id string convert to int64 error: %s", err)
		return
	}

	matchList := reg.Split(hd.deployment, -1)
	afterList := strings.Split(matchList[1], "-")

	hd.serviceID = serviceID
	hd.serviceName = matchList[0]
	hd.phase = afterList[0]
	hd.group = afterList[1]
	log.Infof("get service name: %s phase: %s group: %s", hd.serviceName, hd.phase, hd.group)
}

func (hd *HandleDeployment) scale(replicas int, namespace, deployment string) error {
	// NOTE: 对另一组deployment缩成0
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
