// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/yyjinlong/golib/api"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/service/publish"
)

func Service(r *api.Request) {
	type params struct {
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	serviceName := data.Service
	log.InitFields(log.Fields{"logid": r.TraceID, "service": serviceName})

	sObj := publish.NewService()
	if err := sObj.Handle(serviceName); err != nil {
		log.Errorf("build service failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}
