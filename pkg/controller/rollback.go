// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/yyjinlong/golib/api"
	"github.com/yyjinlong/golib/log"
)

func Rollback(r *api.Request) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Phase    string `form:"phase" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		pid   = data.ID
		phase = data.Phase
		//username = data.Username
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "phase": phase})

	r.ResponseSuccess(nil)
}
