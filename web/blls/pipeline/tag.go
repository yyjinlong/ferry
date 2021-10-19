// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"github.com/gin-gonic/gin"

	"ferry/ops/log"
	"ferry/ops/objects"
	"ferry/web/base"
)

type BuildTag struct{}

func (bt *BuildTag) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID     int64  `form:"pipeline_id" binding:"required"`
		Module string `form:"module" binding:"required"`
		Tag    string `form:"tag" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return "", err
	}
	var (
		pid    = data.ID
		module = data.Module
		tag    = data.Tag
	)
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": pid})

	if err := objects.UpdateTag(pid, module, tag); err != nil {
		return "", err
	}
	log.Infof("module: %s update tag: %s success", module, tag)
	return "", nil
}
