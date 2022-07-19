// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"fmt"

	"nautilus/golib/api"
	"nautilus/golib/log"
	"nautilus/pkg/config"
	"nautilus/pkg/service/publish"
)

func BuildCronjob(r *api.Request) {
	type params struct {
		Namespace string `form:"namespace" binding:"required"`
		Service   string `form:"service" binding:"required"`
		Command   string `form:"command" binding:"required"`
		Schedule  string `form:"schedule" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	log.InitFields(log.Fields{"logid": r.TraceID})

	cron := publish.NewCronjob()
	name, err := cron.Handle(data.Namespace, data.Service, data.Command, data.Schedule)
	if err != nil {
		r.Response(api.Failed, fmt.Sprintf(config.CRON_PUBLISH_ERROR, err), nil)
		return
	}
	r.ResponseSuccess(name)
}
