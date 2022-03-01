// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package view

import (
	"github.com/yyjinlong/golib/api"
	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/bll/pipeline"
)

func CreatePipeline(r *api.Request) {
	type params struct {
		Name       string              `json:"name"`
		Summary    string              `json:"summary"`
		Creator    string              `json:"creator"`
		RD         string              `json:"rd"`
		QA         string              `json:"qa"`
		PM         string              `json:"pm"`
		Service    string              `json:"service"`
		ModuleList []map[string]string `json:"module_list"`
	}

	var data params
	if err := r.BindJSON(&data); err != nil {
		r.Response(api.Failed, err.Error(), nil)
		return
	}

	var (
		name       = data.Name
		summary    = data.Summary
		creator    = data.Creator
		rd         = data.RD
		qa         = data.QA
		pm         = data.PM
		service    = data.Service
		moduleList = data.ModuleList
	)
	log.InitFields(log.Fields{"logid": r.TraceID, "creator": creator, "service": service})
	cp := pipeline.NewCreatePipeline()
	if err := cp.Handle(name, summary, creator, rd, qa, pm, service, moduleList); err != nil {
		log.Errorf("create pipeline failed: %+v", err)
		r.Response(api.Failed, err.Error(), nil)
		return
	}
	r.ResponseSuccess(nil)
}

func ListPipeline(r *api.Request) {

}

func QueryPipeline(r *api.Request) {

}
