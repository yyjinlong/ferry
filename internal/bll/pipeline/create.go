// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/g"
	"nautilus/pkg/log"
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

	var (
		name       = data.Name
		summary    = data.Summary
		moduleList = data.ModuleList
		creator    = data.Creator
		service    = data.Service
	)
	log.InitFields(log.Fields{"logid": r.RequestID, "creator": creator, "service": service})

	if err := b.checkParam(name, summary, moduleList); err != nil {
		return nil, err
	}

	if err := objects.CreatePipeline(name, summary, creator, data.RD, data.QA, data.PM, service, moduleList); err != nil {
		log.Infof("create new pipeline failed: %s", err)
		return nil, fmt.Errorf(DB_CREATE_PIPELINE_ERROR, err)
	}
	log.Infof("create new pipeline success")
	return "", nil
}

func (b *Build) checkParam(name, summary string, moduleList []map[string]string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf(PL_SEGMENT_IS_EMPTY, ONLINE_NAME)
	}
	if strings.TrimSpace(summary) == "" {
		return fmt.Errorf(PL_SEGMENT_IS_EMPTY, ONLINE_DESC)
	}

	for _, item := range moduleList {
		module := item["name"]
		branch := item["branch"]
		if strings.TrimSpace(module) == "" {
			return fmt.Errorf(PL_SEGMENT_IS_EMPTY, MODULE_NAME)
		}
		if strings.TrimSpace(branch) == "" {
			return fmt.Errorf(PL_SEGMENT_IS_EMPTY, BRANCH_NAME)
		}
		if err := b.checkGit(module, branch); err != nil {
			return err
		}
	}
	return nil
}

func (b *Build) checkGit(module, branch string) error {
	codeModule, err := objects.GetCodeModuleInfo(module)
	if err != nil {
		return fmt.Errorf(PL_QUERY_MODULE_ERROR, module)
	}
	param := fmt.Sprintf("git ls-remote --heads %s %s | wc -l", codeModule.ReposAddr, branch)
	log.Infof("git check param: %s", param)
	output, err := g.Cmd("/bin/bash", "-c", param)
	if err != nil {
		log.Errorf("exec git check command: %s error: %s", param, err)
		return fmt.Errorf(PL_EXEC_GIT_CHECK_ERROR, err)
	}
	tmpR := strings.Trim(string(output), "\n")
	newR := strings.TrimSpace(tmpR)
	result, err := strconv.Atoi(newR)
	if err != nil {
		log.Errorf("handle check result covert to int error: %s", err)
		return fmt.Errorf(PL_RESULT_HANDLER_ERROR, err)
	}
	if result == 0 {
		return fmt.Errorf(PL_GIT_CHECK_FAILED)
	}
	return nil
}
