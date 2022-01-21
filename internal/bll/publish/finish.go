// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/log"
)

type Finish struct{}

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
	if errors.Is(err, objects.NotFound) {
		return nil, fmt.Errorf("pipeline_id: %d 不存在!", pid)
	} else if err != nil {
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
