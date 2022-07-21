// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/service/publish"
)

func BuildCronjob(c *gin.Context) {
	type params struct {
		Namespace string `form:"namespace" binding:"required"`
		Service   string `form:"service" binding:"required"`
		Command   string `form:"command" binding:"required"`
		Schedule  string `form:"schedule" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}
	log.InitFields(log.Fields{"logid": r.TraceID})

	cron := publish.NewCronjob()
	name, err := cron.Handle(data.Namespace, data.Service, data.Command, data.Schedule)
	if err != nil {
		r.Response(c, Failed, fmt.Sprintf(config.CRON_PUBLISH_ERROR, err), nil)
		return
	}
	r.ResponseSuccess(c, name)
}
