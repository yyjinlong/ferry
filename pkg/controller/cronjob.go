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

	cron := publish.NewCronjob()
	name, err := cron.Handle(data.Namespace, data.Service, data.Command, data.Schedule)
	if err != nil {
		log.ID(cron.Logid).Errorf("publish cronjob failed: %+v", err)
		Response(c, Failed, fmt.Sprintf(config.CRON_PUBLISH_ERROR, err), nil)
		return
	}
	ResponseSuccess(c, name)
}

func DeleteCronjob(c *gin.Context) {
	type params struct {
		Namespace string `form:"namespace" binding:"required"`
		Service   string `form:"service" binding:"required"`
		JobID     int64  `form:"job_id" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
		return
	}

	cron := publish.NewCronjobDelete()
	if err := cron.Handle(data.Namespace, data.Service, data.JobID); err != nil {
		log.ID(cron.Logid).Errorf("delete cronjob failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}
