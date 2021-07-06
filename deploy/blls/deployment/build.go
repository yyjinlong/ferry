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
	log "github.com/sirupsen/logrus"

	"ferry/deploy/base"
	"ferry/ops/db"
	"ferry/ops/g"
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
	logid      string // 请求id
	phase      string // 当前部署阶段
	group      string // 当前部署组
	deployment string // 当前deployment name
	namespace  string // 当前命名空间
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

	pipelineObj, err := objects.GetPipelineInfo(data.ID)
	if err != nil {
		return nil, err
	}
	imageList, err := objects.FindImageInfo(data.ID)
	if err != nil {
		return nil, err
	}

	b.init(data.ID, r.RequestID, data.Phase, pipelineObj)

	session := db.MEngine.NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		return nil, err
	}

	if err := objects.CreatePhase(session, b.pid, b.phase, db.PHWait); err != nil {
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("create phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.WithFields(log.Fields{
		"logid":       b.logid,
		"pipeline_id": b.pid,
	}).Infof("create phase: %s success", b.phase)

	tpl, err := b.writeYaml(pipelineObj, imageList)
	if err != nil {
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("create yaml error: %s", err)
		return nil, err
	}
	log.WithFields(log.Fields{
		"logid":       b.logid,
		"pipeline_id": b.pid,
	}).Info("create yaml success")

	if b.isFirstDeploy(pipelineObj.Service.Name) {
		if err := b.createDeployment(tpl, b.namespace); err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Info("create deployment success")

	} else {
		if err := b.updateDeployment(tpl, b.namespace, b.deployment); err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Info("update deployment success")
	}

	if err = objects.UpdatePhase(session, b.pid, b.phase, db.PHSuccess, tpl); err != nil {
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("update phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.WithFields(log.Fields{
		"logid":       b.logid,
		"pipeline_id": b.pid,
	}).Infof("update phase: %s success", b.phase)

	if err := session.Commit(); err != nil {
		return nil, err
	}
	return "", nil
}

func (b *Build) init(pipelineID int64, logid, phase string, pipelineObj *db.PipelineQuery) {
	b.pid = pipelineID
	b.logid = logid
	b.phase = phase
	b.namespace = pipelineObj.Namespace.Name

	// 获取部署组
	group := BLUE
	if pipelineObj.Service.OnlineGroup == BLUE {
		group = GREEN
	}
	b.group = group
	log.WithFields(log.Fields{
		"logid":       b.logid,
		"pipeline_id": b.pid,
	}).Infof("fetch current group: %s", group)

	// 获取deployment name
	service := pipelineObj.Service.Name
	b.deployment = fmt.Sprintf("%s-%d-%s-%s", service, b.pid, b.phase, group)
	log.WithFields(log.Fields{
		"logid":       b.logid,
		"pipeline_id": b.pid,
	}).Infof("fetch current deployment name: %s", b.deployment)
}

func (b *Build) writeYaml(pipelineObj *db.PipelineQuery, imageList []db.ImageQuery) (string, error) {
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
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("check first deploy error: %s", err)
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
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("request api-server failed: %s", err)
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
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("parse reqeust deployment result error: %s", err)
		return err
	}

	status, ok := resp["status"].(string)
	if ok && status == "Failure" {
		err := errors.New(resp["message"].(string))
		log.WithFields(log.Fields{
			"logid":       b.logid,
			"pipeline_id": b.pid,
		}).Errorf("request deployment api failed: %s", err)
		return err
	}
	return nil
}
