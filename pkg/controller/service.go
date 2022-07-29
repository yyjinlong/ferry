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
		Response(c, Failed, err.Error(), nil)
		return
	}
	serviceName := data.Service

	sv := publish.NewService()
	if err := sv.Handle(serviceName); err != nil {
		log.Errorf("build service failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
