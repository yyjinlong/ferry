// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package view

import (
	"github.com/yyjinlong/golib/api"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/bll/publish"
)

func Finish(r *api.Request) {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	pid := data.ID
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	finish := publish.NewFinish()
	if err := finish.Handle(pid); err != nil {
		log.Errorf("finish handle failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}
