// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"github.com/gin-gonic/gin"

	"ferry/ops/base"
	"ferry/ops/log"
	"ferry/ops/objects"
)

type BuildTag struct {
}

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
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": data.ID})

	module := data.Module
	tag := data.Tag
	if err := objects.UpdateTag(data.ID, module, tag); err != nil {
		return "", err
	}
	log.Infof("module: %s update tag: %s success", module, tag)
	return "", nil
}
