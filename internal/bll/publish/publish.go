// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"ferry/internal/k8s"
	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/base"
	"ferry/pkg/log"
)

type Deploy struct {
	pid        int64  // pipeline id
	phase      string // 当前部署阶段
	namespace  string // 服务所在命名空间
	deployment string // 服务的deployment名字
	appid      string // 服务的appid
}

func (d *Deploy) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return nil, err
	}
	d.pid = data.ID
	d.phase = data.Phase
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": d.pid, "phase": d.phase})

	// TODO: 建立websocket

	pipeline, err := objects.GetPipelineInfo(d.pid)
	if err != nil {
		return nil, err
	}

	var (
		serviceName = pipeline.Service.Name
		serviceID   = pipeline.Service.ID
		group       = objects.GetDeployGroup(pipeline.Service.OnlineGroup)
	)
	d.namespace = pipeline.Namespace.Name
	d.deployment = objects.GetDeployment(serviceName, serviceID, d.phase, group)
	d.appid = objects.GetAppID(serviceName, serviceID, d.phase)
	log.Infof("[deploy] get current group: %s deployment name: %s", group, d.deployment)

	tpl, err := d.createYaml(pipeline)
	if err != nil {
		log.Errorf("[deploy] generate deployment yaml(%s) error: %+v", d.deployment, err)
		return nil, err
	}
	log.Infof("[deploy] generate deployment yaml(%s) success", d.deployment)

	if err := d.execute(tpl); err != nil {
		return nil, err
	}
	log.Infof("[deploy] pubish deployment: %s to k8s success", d.deployment)

	if err := objects.CreatePhase(d.pid, d.phase, model.PHProcess, tpl); err != nil {
		log.Errorf("[deploy] record deployment: %s to db error: %+v", d.deployment, err)
		return nil, err
	}
	log.Infof("[deploy] record deployment: %s to db success", d.deployment)
	return nil, nil
}

func (d *Deploy) valid() error {
	return nil
}

func (d *Deploy) createYaml(pipeline *model.PipelineQuery) (string, error) {
	imageInfo, err := objects.FindImageInfo(d.pid)
	if err != nil {
		return "", err
	}

	if len(imageInfo) == 0 {
		return "", fmt.Errorf("get image info is empty")
	}
	log.Infof("[deploy] create yaml get image info: %s", imageInfo)

	replicas := pipeline.Service.Replicas
	if d.phase == model.PHASE_SANDBOX {
		// NOTE: 沙盒阶段默认返回1个副本
		replicas = 1
	}

	yaml := &k8s.Yaml{
		Phase:       d.phase,
		Deployment:  d.deployment,
		AppID:       d.appid,
		Namespace:   pipeline.Namespace.Name,
		Service:     pipeline.Service.Name,
		ImageURL:    imageInfo["image_url"],
		ImageTag:    imageInfo["image_tag"],
		Replicas:    replicas,
		QuotaCpu:    pipeline.Service.QuotaCpu,
		QuotaMaxCpu: pipeline.Service.QuotaMaxCpu,
		QuotaMem:    pipeline.Service.QuotaMem,
		QuotaMaxMem: pipeline.Service.QuotaMaxMem,
		VolumeConf:  pipeline.Service.Volume,
		ReserveTime: pipeline.Service.ReserveTime,
	}
	return yaml.Instance()
}

func (d *Deploy) execute(tpl string) error {
	dep := k8s.NewDeployments(d.namespace, d.deployment)
	if !dep.Exist() {
		return dep.Create(tpl)
	}
	return dep.Update(tpl)
}
