// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/objects"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

const (
	TIMEOUT = 5
)

type Build struct {
	pid        int64  // pipeline id
	phase      string // 当前部署阶段
	logid      string // 请求id
	namespace  string // 当前命名空间
	group      string // 当前部署组
	deployment string // 当前deployment name
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.checkParam(c, r.RequestID); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": b.logid, "pipeline_id": b.pid})

	pipelineObj, err := objects.GetPipelineInfo(b.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}
	b.getBaseInfo(pipelineObj)

	if err := objects.CreatePhase(b.pid, b.phase, db.PHWait); err != nil {
		log.Errorf("create phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("create phase: %s success", b.phase)

	tpl, err := b.createYaml(pipelineObj)
	if err != nil {
		log.Errorf("create yaml error: %s", err)
		return nil, err
	}
	log.Info("create yaml success")

	if b.isFirstDeploy(pipelineObj.Service.Name) {
		if err := b.createDeployment(tpl); err != nil {
			return nil, err
		}
		log.Infof("create deployment: %s success", b.deployment)

	} else {
		if err := b.updateDeployment(tpl); err != nil {
			return nil, err
		}
		log.Infof("update deployment: %s success", b.deployment)
	}

	if err = objects.UpdatePhase(b.pid, b.phase, db.PHSuccess, tpl); err != nil {
		log.Errorf("update phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("update phase: %s success", b.phase)
	return "", nil
}

func (b *Build) checkParam(c *gin.Context, logid string) error {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}

	if !g.In(data.Phase, db.PHASE_NAME_LIST) {
		return fmt.Errorf("phase name: %s is not exist", data.Phase)
	}

	b.pid = data.ID
	b.logid = logid
	b.phase = data.Phase
	return nil
}

func (b *Build) getBaseInfo(pipelineObj *db.PipelineQuery) {
	b.namespace = pipelineObj.Namespace.Name
	log.Infof("fetch current namespace: %s", b.namespace)

	group := BLUE
	if pipelineObj.Service.OnlineGroup == BLUE {
		group = GREEN
	}
	b.group = group
	log.Infof("fetch current deploy group: %s", group)

	b.deployment = fmt.Sprintf("%s-%d-%s-%s", pipelineObj.Service.Name, b.pid, b.phase, group)
	log.Infof("fetch current deployment name: %s", b.deployment)
}

func (b *Build) createYaml(pipelineObj *db.PipelineQuery) (string, error) {
	imageList, err := objects.FindImageInfo(b.pid)
	if err != nil {
		return "", err
	}

	yam := &yaml{
		pipelineID:    b.pid,
		phase:         b.phase,
		namespace:     b.namespace,
		deployment:    b.deployment,
		serviceName:   pipelineObj.Service.Name,
		deployPath:    pipelineObj.Service.DeployPath,
		replicas:      b.getReplicas(pipelineObj),
		reserveTime:   pipelineObj.Service.ReserveTime,
		containerConf: pipelineObj.Service.Container,
		volumeConf:    pipelineObj.Service.Volume,
		imageList:     imageList,
	}
	yam.init()
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	return tpl, nil
}

func (b *Build) getReplicas(pipelineObj *db.PipelineQuery) int {
	// NOTE: 沙盒阶段默认返回1个副本
	if b.phase == db.PHASE_SANDBOX {
		return 1
	}
	return pipelineObj.Service.Replicas
}

func (b *Build) isFirstDeploy(service string) bool {
	pipelineList, err := objects.FindPipelineInfo(service)
	if err != nil {
		log.Errorf("check first deploy error: %s", err)
		return false
	}
	if len(pipelineList) == 0 {
		return true
	}
	return false
}

func (b *Build) getBaseURL() string {
	return fmt.Sprintf(g.Config().K8S.Deployment, b.namespace)
}

func (b *Build) createDeployment(tpl string) error {
	url := b.getBaseURL()
	header := map[string]string{"Content-Type": "application/json"}
	body, err := g.Post(url, header, []byte(tpl), TIMEOUT)
	if err != nil {
		log.Errorf("request api-server failed: %s", err)
		return err
	}
	return b.parseResponse(body)
}

func (b *Build) updateDeployment(tpl string) error {
	url := fmt.Sprintf("%s/%s", b.getBaseURL(), b.deployment)
	header := map[string]string{"Content-Type": "application/json"}
	body, err := g.Put(url, header, []byte(tpl), TIMEOUT)
	if err != nil {
		return err
	}
	return b.parseResponse(body)
}

func (b *Build) parseResponse(body string) error {
	resp := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Errorf("json decode http result error: %s", err)
		return err
	}

	status, ok := resp["status"].(string)
	if ok && status == "Failure" {
		err := errors.New(resp["message"].(string))
		log.Errorf("request deployment k8s-api failed: %s", err)
		return err
	}
	return nil
}
