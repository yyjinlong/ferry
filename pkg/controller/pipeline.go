// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"github.com/gin-gonic/gin"

	"nautilus/golib/log"
	"nautilus/pkg/service/pipeline"
)

func CreatePipeline(c *gin.Context) {
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
	if err := c.BindJSON(&data); err != nil {
		Response(c, Failed, err.Error(), nil)
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

	cp := pipeline.NewCreatePipeline()
	if err := cp.Handle(name, summary, creator, rd, qa, pm, service, moduleList); err != nil {
		log.ID(cp.Logid).Errorf("create pipeline failed: %+v", err)
		Response(c, Failed, err.Error(), nil)
		return
	}
	ResponseSuccess(c, nil)
}

func ListPipeline(c *gin.Context) {

}

func QueryPipeline(c *gin.Context) {

}
