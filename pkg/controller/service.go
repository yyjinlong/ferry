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

func Service(c *gin.Context) {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	serviceName := data.Service
	log.InitFields(log.Fields{"logid": r.TraceID, "service": serviceName})

	sObj := publish.NewService()
	if err := sObj.Handle(serviceName); err != nil {
		log.Errorf("build service failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
