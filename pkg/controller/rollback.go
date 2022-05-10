// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"nautilus/golib/api"
	"nautilus/golib/log"
	"nautilus/pkg/service/rollback"
)

func CheckRollback(r *api.Request) {
	type params struct {
		ID    int64  `form:"pipeline_id" binding:"required"`
		Phase string `form:"phase" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		pid   = data.ID
		phase = data.Phase
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "phase": phase})

	r.ResponseSuccess(nil)
}

func Rollback(r *api.Request) {
	type params struct {
		ID       int64  `form:"pipeline_id" binding:"required"`
		Username string `form:"username" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		pid      = data.ID
		username = data.Username
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "username": username})

	ro := rollback.NewRollback()
	if err := ro.Handle(pid, username); err != nil {
		log.Errorf("execute rollback failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}
