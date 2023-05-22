// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/service/publish"
)

func BuildTag(c *gin.Context) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		ResponseFailed(c, err.Error())
		return
	}

	var (
		pid         = data.ID
		serviceName = data.Service
	)

	build := publish.NewBuildTag()
	if err := build.Handle(pid, serviceName); err != nil {
		log.Errorf("build tag failed: %+v", err)
		ResponseFailed(c, err.Error())
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
		ResponseFailed(c, err.Error())
		c.String(http.StatusOK, err.Error())
		return
	}

	var (
		pid    = data.ID
		module = data.Module
		tag    = data.Tag
	)

	receive := publish.NewReceiveTag()
	if err := receive.Handle(pid, module, tag); err != nil {
		log.Errorf("receive tag failed: %+v", err)
		c.String(http.StatusOK, err.Error())
		return
	}
	c.String(http.StatusOK, config.OK)
}
