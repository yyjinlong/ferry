// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"fmt"

	"nautilus/internal/model"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

// GetDeployment 根据服务ID、服务名等创建deployment name
func GetDeployment(serviceName string, serviceID int64, phase, group string) string {
	return fmt.Sprintf("%s-%d-%s-%s", serviceName, serviceID, phase, group)
}

func GetAppID(serviceName string, serviceID int64, phase string) string {
	return fmt.Sprintf("%s-%d-%s", serviceName, serviceID, phase)
}

func GetServiceInfo(name string) (*model.ServiceQuery, error) {
	service := new(model.ServiceQuery)
	session := getServiceSession()
	if has, err := session.Where("service.name = ?", name).Get(service); err != nil {
	} else if !has {
		return nil, NotFound
	}
	return service, nil
}

func GetCodeModules(serviceID int64) ([]model.CodeModule, error) {
	moduleList := make([]model.CodeModule, 0)
	if err := model.SEngine().Where("service_id = ?", serviceID).Find(&moduleList); err != nil {
		return nil, err
	}
	return moduleList, nil
}

func GetPipeline(pipelineID int64) (*model.Pipeline, error) {
	pipeline := new(model.Pipeline)
	if has, err := model.SEngine().ID(pipelineID).Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// GetPipelineInfo 根据pipeline id返回pipeline、namespace、service信息
func GetPipelineInfo(pipelineID int64) (*model.PipelineQuery, error) {
	pq := new(model.PipelineQuery)
	session := getPipelineSession()
	if has, err := session.Where("pipeline.id = ?", pipelineID).Get(pq); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pq, nil
}

// GetServicePipeline 根据服务id返回最近一次的上线信息
func GetServicePipeline(serviceID int64) (*model.PipelineQuery, error) {
	pq := new(model.PipelineQuery)
	session := getPipelineSession()
	if has, err := session.Where("pipeline.service_id = ?", serviceID).Desc("pipeline.id").Get(pq); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pq, nil
}

// FindPhases 根据pipeline id返回对应的阶段
func FindPhases(pipelineID int64) ([]model.PipelinePhase, error) {
	ppList := make([]model.PipelinePhase, 0)
	if err := model.SEngine().Where("pipeline_id=?", pipelineID).Desc("id").Find(&ppList); err != nil {
		return nil, err
	}
	return ppList, nil
}

func CheckPhaseIsDeploy(pipelineID int64, kind, phase string) bool {
	ph := new(model.PipelinePhase)
	if has, err := model.SEngine().Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, phase).Get(ph); err != nil {
		return false
	} else if !has {
		return false
	}
	return true
}

// FindImageInfo 根据pipeline id返回本次构建的镜像信息
func FindImageInfo(pipelineID int64) (map[string]string, error) {
	pi := new(model.ImageQuery)
	session := getImageSession()
	if has, err := session.Where("pipeline.id = ?", pipelineID).Get(pi); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}

	imageInfo := map[string]string{
		"image_url": pi.ImageURL,
		"image_tag": pi.ImageTag,
	}
	return imageInfo, nil
}

// FindPipelineInfo 根据service返回pipeline相关信息
func FindPipelineInfo(service string) ([]model.PipelineQuery, error) {
	pqList := make([]model.PipelineQuery, 0)
	session := getPipelineSession()
	if err := session.Where("service.name = ? AND pipeline.status = ?", service, model.PLSuccess).
		Find(&pqList); err != nil {
		return nil, err
	}
	return pqList, nil
}

func FindUpdateInfo(pipelineID int64) ([]model.UpdateQuery, error) {
	uqList := make([]model.UpdateQuery, 0)
	session := getUpdateSession()
	if err := session.Where("pipeline_update.pipeline_id = ?", pipelineID).Find(&uqList); err != nil {
		return nil, err
	}
	return uqList, nil
}

func GetPhaseInfo(pipelineID int64, kind, phase string) (*model.PipelinePhase, error) {
	phaseObj := new(model.PipelinePhase)
	if has, err := model.SEngine().Where("pipeline_id = ? and kind = ? and name = ?", pipelineID, kind, phase).Get(phaseObj); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return phaseObj, nil
}
