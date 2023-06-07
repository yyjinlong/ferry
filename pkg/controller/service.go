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

func Service(c *gin.Context) {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		ResponseFailed(c, err.Error())
		return
	}
	serviceName := data.Service

	se := publish.NewService()
	if err := se.Handle(serviceName); err != nil {
		log.Errorf("build service failed: %+v", err)
		ResponseFailed(c, err.Error())
		return
	}
	ResponseSuccess(c, nil)
}
