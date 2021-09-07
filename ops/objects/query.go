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
func GetDeployment(serviceName string, serviceID int64, phase, group string) string {
	return fmt.Sprintf("%s-%d-%s-%s", serviceName, serviceID, phase, group)
}

func GetAppID(serviceName string, serviceID int64, phase string) string {
	return fmt.Sprintf("%s-%d-%s", serviceName, serviceID, phase)
}

func GetServiceInfo(name string) (*db.ServiceQuery, error) {
	service := new(db.ServiceQuery)
	if has, err := db.SEngine.Table("service").
		Join("INNER", "namespace", "service.namespace_id = namespace.id").
		Where("service.name = ?", name).Get(service); err != nil {
	} else if !has {
		return nil, fmt.Errorf(NOTFOUND)
	}
	return service, nil
}

func GetCodeModules(serviceID int64) ([]db.CodeModule, error) {
	moduleList := make([]db.CodeModule, 0)
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
	if has, err := db.SEngine.Table("pipeline").
		Join("INNER", "service", "pipeline.service_id = service.id").
		Join("INNER", "namespace", "service.namespace_id = namespace.id").
		Where("pipeline.id = ?", pipelineID).Get(pq); err != nil {
		return nil, err
	} else if !has {
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
func FindImageInfo(pipelineID int64) (map[string]string, error) {
	pi := new(db.ImageQuery)
	if has, err := db.SEngine.Table("pipeline_image").
		Join("INNER", "pipeline", "pipeline.id = pipeline_image.pipeline_id").
		Where("pipeline.id = ?", pipelineID).Get(pi); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf(NOTFOUND)
	}

	imageInfo := map[string]string{
		"image_url": pi.ImageURL,
		"image_tag": pi.ImageTag,
	}
	return imageInfo, nil
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

func FindUpdateInfo(pipelineID int64) ([]db.UpdateQuery, error) {
	uqList := make([]db.UpdateQuery, 0)
	if err := db.SEngine.Table("pipeline_update").
		Join("INNER", "pipeline", "pipeline_update.pipeline_id = pipeline.id").
		Join("INNER", "code_module", "pipeline_update.code_module_id = code_module.id").
		Join("INNER", "service", "pipeline.service_id = service.id").
		Where("pipeline_update.pipeline_id = ?", pipelineID).Find(&uqList); err != nil {
		return nil, err
	}
	return uqList, nil
}
