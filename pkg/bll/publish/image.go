// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/yyjinlong/golib/log"
	"github.com/yyjinlong/golib/rmq"

	"nautilus/pkg/cfg"
	"nautilus/pkg/cm"
	"nautilus/pkg/model"
)

func NewBuildImage() *BuildImage {
	return &BuildImage{}
}

type BuildImage struct{}

func (bi *BuildImage) Handle(pid int64, service string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(cfg.IMG_QUERY_PIPELINE_ERROR, err)
	}

	if cm.Ini(pipeline.Status, []int{model.PLSuccess, model.PLFailed, model.PLRollbackSuccess, model.PLRollbackFailed, model.PLTerminate}) {
		return fmt.Errorf(cfg.IMG_BUILD_FINISHED)
	}

	pipelineImage, err := model.GetImagInfo(pid)
	if err != nil && !errors.Is(err, model.NotFound) {
		return fmt.Errorf(cfg.IMG_QUERY_IS_BUILD_ERROR, err)
	}
	if pipelineImage != nil && cm.Ini(pipelineImage.Status, []int{model.PIProcess, model.PISuccess, model.PIFailed}) {
		return fmt.Errorf(cfg.IMG_QUERY_IMAGE_IS_BUILED)
	}

	updateList, err := model.FindUpdateInfo(pid)
	if err != nil {
		return fmt.Errorf(cfg.IMG_QUERY_UPDATE_ERROR, err)
	}

	language := ""
	builds := make([]map[string]string, 0)
	for _, item := range updateList {
		codeModule, err := model.GetCodeModuleInfoByID(item.CodeModuleID)
		if err != nil {
			return fmt.Errorf(cfg.TAG_QUERY_UPDATE_ERROR, err)
		}
		language = codeModule.Language

		tagInfo := map[string]string{
			"module": codeModule.Name,
			"repo":   codeModule.ReposAddr,
			"tag":    item.CodeTag,
		}
		builds = append(builds, tagInfo)
	}

	image := map[string]interface{}{
		"pid":     pid,
		"type":    language,
		"service": service,
		"build":   builds,
	}
	body, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf(cfg.IMG_BUILD_PARAM_ENCODE_ERROR, err)
	}
	log.Infof("publish build image body: %s", string(body))

	if err := model.CreateImage(pid); err != nil {
		return fmt.Errorf(cfg.IMG_CREATE_IMAGE_INFO_ERROR, err)
	}

	mq, err := rmq.NewRabbitMQ(
		cfg.Config().RabbitMQ.Address,
		cfg.Config().RabbitMQ.Exchange,
		cfg.Config().RabbitMQ.Queue,
		cfg.Config().RabbitMQ.RoutingKey)
	if err != nil {
		return fmt.Errorf(cfg.IMG_SEND_BUILD_TO_MQ_FAILED, err)
	}
	mq.Publish(string(body))
	log.Infof("publish build image info to rabbitmq success.")
	return nil
}
