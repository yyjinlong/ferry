// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"fmt"

	"ferry/ops/db"
)

const (
	BLUE     = "blue"
	GREEN    = "green"
	NOTFOUND = "no data not found."
)

// GetDeployGroup 获取当前部署组
func GetDeployGroup(onlineGroup string) string {
	group := BLUE
	if onlineGroup == BLUE {
		group = GREEN
	}
	return group
}

// GetDeployment 根据服务ID、服务名等创建deployment name
func GetDeployment(serviceID int64, service, phase, group string) string {
	return fmt.Sprintf("%s-%d-%s-%s", service, serviceID, phase, group)
}

func GetService(name string) (*db.Service, error) {
	service := new(db.Service)
	if has, err := db.SEngine.Where("name=?", name).Get(service); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf(NOTFOUND)
	}
	return service, nil
}

func GetModules(serviceID int64) ([]db.Module, error) {
	moduleList := make([]db.Module, 0)
	if err := db.SEngine.Where("service_id=?", serviceID).Find(&moduleList); err != nil {
		return nil, err
	}
	return moduleList, nil
}

func GetPipeline(pipelineID int64) (*db.Pipeline, error) {
	pipeline := new(db.Pipeline)
	if has, err := db.SEngine.ID(pipelineID).Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf(NOTFOUND)
	}
	return pipeline, nil
}

// GetPipelineInfo 根据pipeline id返回pipeline、namespace、service信息
func GetPipelineInfo(pipelineID int64) (*db.PipelineQuery, error) {
	pq := new(db.PipelineQuery)
	has, err := db.SEngine.Table("pipeline").
		Join("INNER", "service", "pipeline.service_id = service.id").
		Join("INNER", "namespace", "service.namespace_id = namespace.id").
		Where("pipeline.id = ?", pipelineID).Get(pq)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf(NOTFOUND)
	}
	return pq, nil
}

// FindPhases 根据pipeline id返回对应的阶段
func FindPhases(pipelineID int64) ([]db.PipelinePhase, error) {
	ppList := make([]db.PipelinePhase, 0)
	if err := db.SEngine.Where("pipeline_id=?", pipelineID).Find(&ppList); err != nil {
		return nil, err
	}
	return ppList, nil
}

// FindImageInfo 根据pipeline id返回本次构建的镜像信息
func FindImageInfo(pipelineID int64) ([]map[string]string, error) {
	piList := make([]db.ImageQuery, 0)
	if err := db.SEngine.Table("pipeline_image").
		Join("INNER", "pipeline", "pipeline_image.pipeline_id = pipeline.id").
		Join("INNER", "module", "pipeline_image.module_id = module.id").
		Where("pipeline_image.pipeline_id = ?", pipelineID).
		Find(&piList); err != nil {
		return nil, err
	}

	imageList := make([]map[string]string, 0)
	for _, model := range piList {
		imageInfo := map[string]string{
			"module_name": model.Module.Name,
			"image_url":   model.PipelineImage.ImageURL,
			"image_tag":   model.PipelineImage.ImageTag,
		}
		imageList = append(imageList, imageInfo)
	}
	return imageList, nil
}

// FindPipelineInfo 根据service返回pipeline相关信息
func FindPipelineInfo(service string) ([]db.PipelineQuery, error) {
	pqList := make([]db.PipelineQuery, 0)
	if err := db.SEngine.Table("pipeline").
		Join("INNER", "service", "pipeline.service_id = service.id").
		Join("INNER", "namespace", "service.namespace_id = namespace.id").
		Where("service.name = ? AND pipeline.status = ?", service, db.PLSuccess).
		Find(&pqList); err != nil {
		return nil, err
	}
	return pqList, nil
}
