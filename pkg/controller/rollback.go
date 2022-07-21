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

	//var (
	//	pid   = data.ID
	//	phase = data.Phase
	//)
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

	ro := rollback.NewRollback()
	if err := ro.Handle(pid, username); err != nil {
		log.ID(ro.Logid).Errorf("execute rollback failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
