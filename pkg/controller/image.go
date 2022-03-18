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

func BuildImage(r *api.Request) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		pid     = data.ID
		service = data.Service
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	image := publish.NewBuildImage()
	if err := image.Handle(pid, service); err != nil {
		log.Errorf("build image pre handle failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}
