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

func BuildTag(c *gin.Context) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		pid         = data.ID
		serviceName = data.Service
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	build := publish.NewBuildTag()
	if err := build.Handle(pid, serviceName); err != nil {
		log.Errorf("build tag failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}

func ReceiveTag(c *gin.Context) {
	type params struct {
		ID     int64  `form:"taskid" binding:"required"`
		Module string `form:"module" binding:"required"`
		Tag    string `form:"tag" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	var (
		pid    = data.ID
		module = data.Module
		tag    = data.Tag
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	receive := publish.NewReceiveTag()
	if err := receive.Handle(pid, module, tag); err != nil {
		log.Errorf("receive tag failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
