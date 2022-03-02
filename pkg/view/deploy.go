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

func Deploy(r *api.Request) {
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
	pid := data.ID
	phase := data.Phase
	username := data.Username
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid, "phase": phase})

	dep := publish.NewDeploy()
	if err := dep.Handle(pid, phase, username); err != nil {
		log.Errorf("build deployment failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}
