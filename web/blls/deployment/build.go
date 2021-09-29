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
	pid   int64  // pipeline id
	phase string // 当前部署阶段
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.validate(c); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": b.pid, "phase": b.phase})

	pipeline, err := objects.GetPipelineInfo(b.pid)
	if err != nil {
		return nil, err
	}

	var (
		namespace   = pipeline.Namespace.Name
		serviceName = pipeline.Service.Name
		serviceID   = pipeline.Service.ID
	)

	var (
		group      = objects.GetDeployGroup(pipeline.Service.OnlineGroup)
		deployment = objects.GetDeployment(serviceName, serviceID, b.phase, group)
		appid      = objects.GetAppID(serviceName, serviceID, b.phase)
	)
	log.Infof("get current group: %s deployment: %s", group, deployment)

	tpl, err := b.createYaml(pipeline, deployment, appid)
	if err != nil {
		log.Errorf("create yaml error: %+v", err)
		return nil, err
	}
	log.Infof("create deployment: %s yaml success", deployment)

	if err := b.publish(namespace, deployment, tpl); err != nil {
		return nil, err
	}
	log.Infof("pubish deployment: %s success", deployment)

	if err := objects.CreatePhase(b.pid, b.phase, db.PHProcess); err != nil {
		log.Errorf("create db record error: %+v", err)
		return nil, err
	}
	log.Info("create db record success")
	return "", nil
}

func (b *Build) validate(c *gin.Context) error {
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
	b.phase = data.Phase
	return nil
}

func (b *Build) createYaml(pipeline *db.PipelineQuery, deployment, appid string) (string, error) {
	imageInfo, err := objects.FindImageInfo(b.pid)
	if err != nil {
		return "", err
	}

	if len(imageInfo) == 0 {
		return "", fmt.Errorf("get image info is empty")
	}
	log.Infof("create yaml get image info: %s", imageInfo)

	replicas := pipeline.Service.Replicas
	if b.phase == db.PHASE_SANDBOX {
		// NOTE: 沙盒阶段默认返回1个副本
		replicas = 1
	}

	yam := &yaml{
		pipelineID:  b.pid,
		phase:       b.phase,
		deployment:  deployment,
		appid:       appid,
		namespace:   pipeline.Namespace.Name,
		service:     pipeline.Service.Name,
		imageURL:    imageInfo["image_url"],
		imageTag:    imageInfo["image_tag"],
		replicas:    replicas,
		quotaCpu:    pipeline.Service.QuotaCpu,
		quotaMaxCpu: pipeline.Service.QuotaMaxCpu,
		quotaMem:    pipeline.Service.QuotaMem,
		quotaMaxMem: pipeline.Service.QuotaMaxMem,
		volumeConf:  pipeline.Service.Volume,
		reserveTime: pipeline.Service.ReserveTime,
	}
	return yam.instance()
}

func (b *Build) publish(namespace, deployment, tpl string) error {
	dep := newDeployments()
	if !dep.exist(namespace, deployment) {
		return dep.create(namespace, tpl)
	}
	return dep.update(namespace, deployment, tpl)
}
