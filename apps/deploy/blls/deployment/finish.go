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
	"ferry/ops/log"
	"ferry/ops/objects"
)

type Finish struct {
	pid int64
}

func (f *Finish) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	if err := f.checkParam(c, r.RequestID); err != nil {
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": f.pid})

	pipeline, err := objects.GetPipelineInfo(f.pid)
	if err != nil {
		log.Errorf("get pipeline info error: %s", err)
		return nil, err
	}

	if !f.clearOld(pipeline) {
		return nil, fmt.Errorf("old group deployment scale to 0 failed")
	}

	if err := f.setOnline(pipeline); err != nil {
		return nil, err
	}
	return "", nil
}

func (f *Finish) checkParam(c *gin.Context, logid string) error {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return err
	}
	f.pid = data.ID
	return nil
}

func (f *Finish) clearOld(pipeline *db.PipelineQuery) bool {
	namespace := pipeline.Namespace.Name

	// NOTE: 在确认时, 原有表记录的组则变为待下线组
	offlineGroup := pipeline.Service.OnlineGroup
	log.Infof("get current clear offline group: %s", offlineGroup)
	if offlineGroup == "none" {
		return true
	}

	dep := newDeployments()
	for _, phase := range db.PHASE_NAME_LIST {
		deployment := objects.GetDeployment(pipeline.Service.Name, pipeline.Service.ID, phase, offlineGroup)
		if dep.exist(namespace, deployment) {
			if err := dep.scale(0, namespace, deployment); err != nil {
				log.Errorf("scale deployment: %s replicas: 0 error: %s", deployment, err)
				return false
			}
			log.Infof("scale deployment: %s replicas: 0 success", deployment)
		}
	}
	return true
}

func (f *Finish) setOnline(pipeline *db.PipelineQuery) error {
	group := objects.GetDeployGroup(pipeline.Service.OnlineGroup)
	log.Infof("get current online group: %s", group)

	if err := objects.UpdateGroup(f.pid, pipeline.Service.Name, group); err != nil {
		log.Errorf("set current group: %s online error: %s", group, err)
	}
	log.Infof("set current group: %s online success.", group)
	return nil
}
