// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/base"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

// TODO: 返回一个公共方法，创建一个deployment

type Build struct {
	pid   int64  // pipeline id
	phase string // 当前部署阶段
}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := b.valid(c); err != nil {
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
		log.Errorf("create deployment: %s yaml error: %+v", deployment, err)
		return nil, err
	}
	log.Infof("create deployment: %s yaml success", deployment)

	if err := b.publish(namespace, deployment, tpl); err != nil {
		return nil, err
	}
	log.Infof("pubish deployment: %s to k8s success", deployment)

	if err := objects.CreatePhase(b.pid, b.phase, model.PHProcess, tpl); err != nil {
		log.Errorf("record deployment: %s to db error: %+v", deployment, err)
		return nil, err
	}
	log.Infof("record deployment: %s to db success", deployment)
	return "", nil
}

func (b *Build) valid(c *gin.Context) error {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}

	if !g.In(data.Phase, model.PHASE_NAME_LIST) {
		return fmt.Errorf("phase name: %s is not exist", data.Phase)
	}

	b.pid = data.ID
	b.phase = data.Phase
	return nil
}

func (b *Build) createYaml(pipeline *model.PipelineQuery, deployment, appid string) (string, error) {
	imageInfo, err := objects.FindImageInfo(b.pid)
	if err != nil {
		return "", err
	}

	if len(imageInfo) == 0 {
		return "", fmt.Errorf("get image info is empty")
	}
	log.Infof("create yaml get image info: %s", imageInfo)

	replicas := pipeline.Service.Replicas
	if b.phase == model.PHASE_SANDBOX {
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
	dep := NewDeployments(namespace, deployment)
	if !dep.Exist() {
		return dep.Create(tpl)
	}
	return dep.Update(tpl)
}
