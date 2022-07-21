// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/gin-gonic/gin"

	"nautilus/golib/log"
	"nautilus/pkg/service/rollback"
)

func CheckRollback(c *gin.Context) {
	type params struct {
		ID    int64  `form:"pipeline_id" binding:"required"`
		Phase string `form:"phase" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		pid   = data.ID
		phase = data.Phase
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "phase": phase})

	ResponseSuccess(c, nil)
}

func Rollback(c *gin.Context) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		pid      = data.ID
		username = data.Username
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "username": username})

	ro := rollback.NewRollback()
	if err := ro.Handle(pid, username); err != nil {
		log.Errorf("execute rollback failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
