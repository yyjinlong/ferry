// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package deployment

import (
	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/log"
	"ferry/ops/objects"
)

type Finish struct {
}

func (f *Finish) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return nil, err
	}

	pid := data.ID
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": pid})

	pipeline, err := objects.GetPipelineInfo(pid)
	if err != nil {
		return nil, err
	}

	group := objects.GetDeployGroup(pipeline.Service.OnlineGroup)
	log.Infof("get current online group: %s", group)

	if err := objects.UpdateGroup(pid, pipeline.Service.Name, group); err != nil {
		log.Errorf("set current group: %s online error: %s", group, err)
	}
	log.Infof("set current group: %s online success.", group)
	return "", nil
}
