// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
)

func NewBuildImage(pid int64, service string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_PIPELINE_ERROR, err)
	}

	statusList := []int{
		model.PLSuccess,
		model.PLFailed,
		model.PLRollbackSuccess,
		model.PLRollbackFailed,
		model.PLTerminate,
	}
	if cm.Ini(pipeline.Status, statusList) {
		return fmt.Errorf(config.IMG_BUILD_FINISHED)
	}

	if err := model.CreatePhase(pid, model.KIND_DEPLOY, model.PHASE_IMAGE, model.PHProcess); err != nil {
		log.Errorf("create pipeline: %d image phase error: %s", pid, err)
		return err
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_UPDATE_ERROR, err)
	}

	_, curPath, _, _ := runtime.Caller(1)
	var (
		mainPath   = filepath.Dir(filepath.Dir(filepath.Dir(curPath)))
		scriptPath = filepath.Join(mainPath, "script")
		changes    []string
		retains    []string
	)

	for _, item := range updateList {
		module := item.CodeModule
		if err := model.CreateOrUpdatePipelineImage(pid, service, module, "", ""); err != nil {
			return err
		}

		output := ""
		param := fmt.Sprintf("%s/makeimg -s %s -m %s -p %s -i %d", scriptPath, service, module, item.CodePkg, pid)
		log.Infof("makeimg command: %s", param)
		if err := CallRealtimeOut(param, &output, nil); err != nil {
			return fmt.Errorf(config.IMG_BUILD_FAILED)
		}
		changes = append(changes, item.CodeModule)
		fmt.Println(output)
	}

	// 获取未变更的模块(服务所有模块-当前变更的模块)
	totals, err := model.FindServiceCodeModules(service)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_MODULE_BINDING_ERROR, err)
	}

	for _, item := range totals {
		codeModule := item.CodeModule.Name
		if cm.In(codeModule, changes) {
			continue
		}
		retains = append(retains, codeModule)
	}
	log.Infof("build image pipeline: %d fetch unchange code modules: %v", pid, retains)

	for _, codeModule := range retains {
		image, err := model.QueryLatestSuccessModuleImage(service, codeModule)
		if err != nil {
			log.Errorf(config.DB_IMAGE_CREATE_OR_UPDATE_ERROR, err)
			return err
		}
		imageURL := image.ImageURL
		imageTag := image.ImageTag

		if err := model.CreateOrUpdatePipelineImage(pid, service, codeModule, imageURL, imageTag); err != nil {
			return err
		}
		log.Infof("build image pipeline: %d record latest module: %s image: %s:%s success", pid, codeModule, imageURL, imageTag)
	}

	if err := model.UpdatePhase(pid, model.KIND_DEPLOY, model.PHASE_IMAGE, model.PHSuccess); err != nil {
		log.Errorf("update pipeline: %d image phase error: %s", pid, err)
	}
	return nil
}

func UpdateImageInfo(pid int64, module, imageURL, imageTag string) error {
	return model.UpdateImage(pid, module, imageURL, imageTag)
}
