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

func BuildImage(c *gin.Context) {
	type params struct {
		ID      *int64 `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		ResponseFailed(c, err.Error())
		return
	}

	var (
		pid     = *data.ID
		service = data.Service
	)

	if err := publish.NewBuildImage(pid, service); err != nil {
		log.Errorf("build image pre handle failed: %+v", err)
		ResponseFailed(c, err.Error())
		return
	}
	ResponseSuccess(c, nil)
}

func UpdateImage(c *gin.Context) {
	type params struct {
		ID       *int64 `form:"taskid" binding:"required"`
		Module   string `form:"module" binding:"required"`
		ImageURL string `form:"image_url" binding:"required"`
		ImageTag string `form:"image_tag" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}

	if err := publish.UpdateImageInfo(*data.ID, data.Module, data.ImageURL, data.ImageTag); err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	c.String(http.StatusOK, config.OK)
}
