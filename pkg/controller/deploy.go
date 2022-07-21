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

func Deploy(c *gin.Context) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		pid      = data.ID
		phase    = data.Phase
		username = data.Username
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "phase": phase})

	dep := publish.NewDeploy()
	if err := dep.Handle(pid, phase, username); err != nil {
		log.Errorf("build deployment failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
