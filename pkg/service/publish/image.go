// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
	"nautilus/pkg/util/rmq"
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

	pipelineImage, err := model.GetImagInfo(pid)
	if err != nil && !errors.Is(err, model.NotFound) {
		return fmt.Errorf(config.IMG_QUERY_IS_BUILD_ERROR, err)
	}
	if pipelineImage != nil && cm.Ini(pipelineImage.Status, []int{model.PIProcess, model.PISuccess, model.PIFailed}) {
		return fmt.Errorf(config.IMG_QUERY_IMAGE_IS_BUILED)
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(config.IMG_QUERY_UPDATE_ERROR, err)
	}

	language := ""
	builds := make([]config.ModuleInfo, 0)
	for _, item := range updateList {
		codeModule, err := model.GetCodeModuleInfoByID(item.CodeModuleID)
		if err != nil {
			return fmt.Errorf(config.TAG_QUERY_UPDATE_ERROR, err)
		}
		language = codeModule.Language

		builds = append(builds, config.ModuleInfo{
			Module: codeModule.Name,
			Repo:   codeModule.ReposAddr,
			Tag:    item.CodeTag,
		})
	}

	image := config.Image{
		PID:     pid,
		Type:    language,
		Service: service,
		Build:   builds,
	}
	body, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf(config.IMG_BUILD_PARAM_ENCODE_ERROR, err)
	}
	log.Infof("publish build image body: %s", string(body))

	if err := model.CreateImage(pid); err != nil {
		return fmt.Errorf(config.IMG_CREATE_IMAGE_INFO_ERROR, err)
	}

	mq, err := rmq.NewRabbitMQ(
		config.Config().RabbitMQ.Address,
		config.Config().RabbitMQ.Exchange,
		config.Config().RabbitMQ.Queue,
		config.Config().RabbitMQ.RoutingKey)
	if err != nil {
		return fmt.Errorf(config.IMG_SEND_BUILD_TO_MQ_FAILED, err)
	}
	mq.Publish(string(body))
	log.Infof("publish build image info to rabbitmq success")
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
