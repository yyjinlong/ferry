// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/db"
	"ferry/ops/g"
	"ferry/ops/log"
	"ferry/ops/objects"
)

type Build struct {
	pid        int64  // pipeline id
	phase      string // 当前部署阶段
	logid      string // 请求id
	namespace  string // 当前命名空间
	service    string // 当前服务名
	group      string // 当前部署组
	deployment string // 当前deployment name
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.checkParam(c, r.RequestID); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": b.logid, "pipeline_id": b.pid})

	pipeline, err := objects.GetPipelineInfo(b.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}

	b.group = objects.GetDeployGroup(pipeline.Service.OnlineGroup)
	log.Infof("fetch current deploy group: %s", b.group)

	b.namespace = pipeline.Namespace.Name
	b.service = pipeline.Service.Name
	b.deployment = objects.GetDeployment(pipeline.Service.ID, b.service, b.phase, b.group)
	log.Infof("fetch current deployment name: %s", b.deployment)

	if err := objects.CreatePhase(b.pid, b.phase, db.PHWait); err != nil {
		log.Errorf("create %s phase error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("create %s phase success", b.phase)

	tpl, err := b.createYaml(pipeline)
	if err != nil {
		log.Errorf("create yaml error: %s", err)
		return nil, err
	}
	log.Info("create yaml success")

	dep := newDeployments()
	if !dep.exist(b.namespace, b.deployment) {
		if err := dep.create(b.namespace, tpl); err != nil {
			return nil, err
		}
		log.Infof("create deployment: %s success", b.deployment)

	} else {
		if err := dep.update(b.namespace, b.deployment, tpl); err != nil {
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

func (b *Build) createYaml(pipeline *db.PipelineQuery) (string, error) {
	imageList, err := objects.FindImageInfo(b.pid)
	if err != nil {
		return "", err
	}
	log.Infof("create yaml get image list: %s", imageList)
	if len(imageList) == 0 {
		return "", fmt.Errorf("get image list is empty")
	}

	yam := &yaml{
		pipelineID:    b.pid,
		phase:         b.phase,
		namespace:     b.namespace,
		deployment:    b.deployment,
		serviceName:   b.service,
		deployPath:    pipeline.Service.DeployPath,
		replicas:      b.getReplicas(pipeline),
		reserveTime:   pipeline.Service.ReserveTime,
		containerConf: pipeline.Service.Container,
		volumeConf:    pipeline.Service.Volume,
		imageList:     imageList,
	}
	yam.init()
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	return tpl, nil
}

func (b *Build) getReplicas(pipeline *db.PipelineQuery) int {
	// NOTE: 沙盒阶段默认返回1个副本
	if b.phase == db.PHASE_SANDBOX {
		return 1
	}
	return pipeline.Service.Replicas
}
