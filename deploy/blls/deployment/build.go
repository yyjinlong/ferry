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

	"ferry/deploy/base"
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
	pid        int64      // pipeline id
	phase      string     // 当前部署阶段
	logid      string     // 请求id
	group      string     // 当前部署组
	deployment string     // 当前deployment name
	namespace  string     // 当前命名空间
	fields     log.Fields // 日志公共部分
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return nil, err
	}

	b.init(data.ID, r.RequestID, data.Phase)
	pipelineObj, err := b.getBaseInfo()
	if err != nil {
		return nil, err
	}

	session := db.MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	if err := objects.CreatePhase(session, b.pid, b.phase, db.PHWait); err != nil {
		log.Errorf("create phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("create phase: %s success", b.phase)

	tpl, err := b.writeYaml(pipelineObj)
	if err != nil {
		log.Errorf("create yaml error: %s", err)
		return nil, err
	}
	log.Info("create yaml success")

	if b.isFirstDeploy(pipelineObj.Service.Name) {
		if err := b.createDeployment(tpl, b.namespace); err != nil {
			return nil, err
		}
		log.Info("create deployment success")

	} else {
		if err := b.updateDeployment(tpl, b.namespace, b.deployment); err != nil {
			return nil, err
		}
		log.Info("update deployment success")
	}

	if err = objects.UpdatePhase(session, b.pid, b.phase, db.PHSuccess, tpl); err != nil {
		log.Errorf("update phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("update phase: %s success", b.phase)

	if err := session.Commit(); err != nil {
		return nil, err
	}
	return "", nil
}

func (b *Build) init(pipelineID int64, logid, phase string) {
	b.pid = pipelineID
	b.logid = logid
	b.phase = phase
	log.InitFields(log.Fields{"logid": logid, "pipeline_id": pipelineID})
}

func (b *Build) getBaseInfo() (*db.PipelineQuery, error) {
	pipelineObj, err := objects.GetPipelineInfo(b.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}
	b.namespace = pipelineObj.Namespace.Name

	group := BLUE
	if pipelineObj.Service.OnlineGroup == BLUE {
		group = GREEN
	}
	b.group = group
	log.Infof("fetch current deploy group: %s", group)

	b.deployment = fmt.Sprintf("%s-%d-%s-%s", pipelineObj.Service.Name, b.pid, b.phase, group)
	log.Infof("fetch current deployment name: %s", b.deployment)
	return pipelineObj, nil
}

func (b *Build) writeYaml(pipelineObj *db.PipelineQuery) (string, error) {
	imageList, err := objects.FindImageInfo(b.pid)
	if err != nil {
		return "", err
	}

	yam := &yaml{
		pipelineID:     b.pid,
		logid:          b.logid,
		phase:          b.phase,
		group:          b.group,
		namespace:      pipelineObj.Namespace.Name,
		deploymentName: b.deployment,
		serviceObj:     &pipelineObj.Service,
		imageList:      imageList,
	}
	yam.init()
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	return tpl, nil
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

func (b *Build) createDeployment(tpl, namespace string) error {
	url := fmt.Sprintf(g.Config().K8S.Deployment, namespace)
	header := make(map[string]string)
	header["Context-Type"] = "application/json"
	body, err := g.Post(url, header, []byte(tpl), TIMEOUT)
	if err != nil {
		log.Errorf("request api-server failed: %s", err)
		return err
	}
	return b.analysis(body)
}

func (b *Build) updateDeployment(tpl, namespace, deploymentName string) error {
	baseURL := fmt.Sprintf(g.Config().K8S.Deployment, namespace)
	url := fmt.Sprintf("%s/%s", baseURL, deploymentName)

	header := make(map[string]string)
	header["Context-Type"] = "application/json"
	body, err := g.Put(url, header, []byte(tpl), TIMEOUT)
	if err != nil {
		return err
	}
	return b.analysis(body)
}

func (b *Build) analysis(body string) error {
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
