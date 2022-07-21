// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/gin-gonic/gin"

	"nautilus/golib/log"
	"nautilus/pkg/service/publish"
)

func Finish(c *gin.Context) {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	pid := data.ID
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	finish := publish.NewFinish()
	if err := finish.Handle(pid); err != nil {
		log.Errorf("finish handle failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
