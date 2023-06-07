// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
)

func NewCreatePipeline() *CreatePipeline {
	return &CreatePipeline{}
}

type CreatePipeline struct{}

func (cp *CreatePipeline) Handle(name, summary, creator, rd, qa, pm, service string, moduleList []map[string]string) error {
	if err := cp.checkParam(name, summary, moduleList); err != nil {
		return err
	}

	if err := model.CreatePipeline(name, summary, creator, rd, qa, pm, service, moduleList); err != nil {
		return fmt.Errorf(config.PL_CREATE_PIPELINE_ERROR, err)
	}
	log.Infof("create pipeline success")
	return nil
}

func (cp *CreatePipeline) checkParam(name, summary string, moduleList []map[string]string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf(config.PL_SEGMENT_IS_EMPTY, config.ONLINE_NAME)
	}
	if strings.TrimSpace(summary) == "" {
		return fmt.Errorf(config.PL_SEGMENT_IS_EMPTY, config.ONLINE_DESC)
	}

	for _, item := range moduleList {
		module := item["name"]
		branch := item["branch"]
		if strings.TrimSpace(module) == "" {
			return fmt.Errorf(config.PL_SEGMENT_IS_EMPTY, config.MODULE_NAME)
		}
		if strings.TrimSpace(branch) == "" {
			return fmt.Errorf(config.PL_SEGMENT_IS_EMPTY, config.BRANCH_NAME)
		}
		if err := cp.checkGit(module, branch); err != nil {
			return err
		}
	}
	return nil
}

func (cp *CreatePipeline) checkGit(module, branch string) error {
	codeModule, err := model.GetCodeModuleInfo(module)
	if err != nil {
		return fmt.Errorf(config.PL_QUERY_MODULE_ERROR, module)
	}
	param := fmt.Sprintf("git ls-remote --heads %s %s | wc -l", codeModule.RepoAddr, branch)
	log.Infof("git check param: %s", param)
	output, err := cm.Call(param)
	if err != nil {
		log.Errorf("exec git check command: %s error: %s", param, err)
		return fmt.Errorf(config.PL_EXEC_GIT_CHECK_ERROR, err)
	}
	tmpR := strings.Trim(string(output), "\n")
	newR := strings.TrimSpace(tmpR)
	result, err := strconv.Atoi(newR)
	if err != nil {
		log.Errorf("handle check result covert to int error: %s", err)
		return fmt.Errorf(config.PL_RESULT_HANDLER_ERROR, err)
	}
	if result == 0 {
		return fmt.Errorf(config.PL_GIT_CHECK_FAILED)
	}
	return nil
}
