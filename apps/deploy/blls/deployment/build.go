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

func (b *Build) createYaml(pipeline *db.PipelineQuery, deployment string) (string, error) {
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
	tpl, err := yam.instance()
	if err != nil {
		return "", err
	}
	return tpl, nil
}

func (b *Build) publish(namespace, deployment, tpl string) error {
	dep := newDeployments()
	if !dep.exist(namespace, deployment) {
		if err := dep.create(namespace, tpl); err != nil {
			return err
		}
		log.Infof("create deployment: %s success", deployment)

	} else {
		if err := dep.update(namespace, deployment, tpl); err != nil {
			return err
		}
		log.Infof("update deployment: %s success", deployment)
	}
	return nil
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.validate(c); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": b.pid})

	pipeline, err := objects.GetPipelineInfo(b.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}

	group := objects.GetDeployGroup(pipeline.Service.OnlineGroup)
	log.Infof("get current deploy group: %s", group)

	deployment := objects.GetDeployment(pipeline.Service.Name, pipeline.Service.ID, b.phase, group)
	log.Infof("get current deployment name: %s", deployment)

	if err := objects.CreatePhase(b.pid, b.phase, db.PHWait); err != nil {
		log.Errorf("create %s phase error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("create %s phase success", b.phase)

	tpl, err := b.createYaml(pipeline, deployment)
	if err != nil {
		log.Errorf("create yaml error: %s", err)
		return nil, err
	}
	log.Infof("create %s phase yaml success", b.phase)

	if err := b.publish(pipeline.Namespace.Name, deployment, tpl); err != nil {
		log.Errorf("publish deployment failed: %s", err)
		return nil, err
	}

	if err = objects.UpdatePhase(b.pid, b.phase, db.PHSuccess, tpl); err != nil {
		log.Errorf("update phase: %s error: %s", b.phase, err)
		return nil, err
	}
	log.Infof("update phase: %s success", b.phase)
	return "", nil
}
