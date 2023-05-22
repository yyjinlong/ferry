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

func NewBuildImage() *BuildImage {
	return &BuildImage{}
}

type BuildImage struct{}

func (bi *BuildImage) Handle(pid int64, service string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_PIPELINE_ERROR, err)
	}

	if err := bi.checkStatus(pipeline.Status); err != nil {
		return err
	}

	svc, err := model.GetServiceInfo(service)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_SERVICE_ERROR, err)
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_UPDATE_ERROR, err)
	}

	_, curPath, _, _ := runtime.Caller(1)
	var (
		mainPath   = filepath.Dir(filepath.Dir(filepath.Dir(curPath)))
		scriptPath = filepath.Join(mainPath, "script")
	)

	for _, item := range updateList {
		if err := model.CreateImage(pid, item.CodeModule); err != nil {
			return fmt.Errorf(config.IMG_CREATE_IMAGE_INFO_ERROR, err)
		}

		codeModule, err := model.GetCodeModuleInfo(item.CodeModule)
		if err != nil {
			return fmt.Errorf(config.TAG_QUERY_UPDATE_ERROR, err)
		}
		lang := codeModule.Language
		repo := codeModule.ReposAddr

		param := fmt.Sprintf("%s/makeimg -s %s -m %s -l %s -a %s -t %s -u %s -i %d",
			scriptPath, service, item.CodeModule, lang, repo, item.CodeTag, svc.ImageAddr, pid)
		log.Infof("makeimg command: %s", param)
		if !cm.CallRealtimeOut(param) {
			return fmt.Errorf(config.IMG_BUILD_FAILED)
		}
	}
	return nil
}

func (bi *BuildImage) checkStatus(status int) error {
	statusList := []int{
		model.PLSuccess,
		model.PLFailed,
		model.PLRollbackSuccess,
		model.PLRollbackFailed,
		model.PLTerminate,
	}
	if cm.Ini(status, statusList) {
		return fmt.Errorf(config.IMG_BUILD_FINISHED)
	}
	return nil
}

func UpdateImageInfo(pid int64, module, imageURL, imageTag string) error {
	return model.UpdateImage(pid, module, imageURL, imageTag)
}
