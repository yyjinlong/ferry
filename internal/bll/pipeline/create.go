// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"github.com/gin-gonic/gin"

	"ferry/internal/objects"
	"ferry/pkg/base"
	"ferry/pkg/log"
)

type Build struct{}

func (b *Build) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
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
		return nil, err
	}
	log.InitFields(log.Fields{"logid": r.RequestID})

	if err := objects.CreatePipeline(data.Name, data.Summary, data.Creator,
		data.RD, data.QA, data.PM, data.Service, data.ModuleList); err != nil {
		log.Infof("create new pipeline error: %s", err)
		return nil, err
	}
	log.Infof("create new pipeline success.")
	return "", nil
}

func (b *Build) valid() {

}
