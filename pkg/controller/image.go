// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"nautilus/pkg/service/publish"
)

func BuildImage(c *gin.Context) {
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
		pid     = data.ID
		service = data.Service
	)

	image := publish.NewBuildImage()
	if err := image.Handle(pid, service); err != nil {
		log.Errorf("build image pre handle failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
