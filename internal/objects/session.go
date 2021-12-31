// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"xorm.io/xorm"

	"ferry/internal/model"
)

func getServiceSession() *xorm.Session {
	return model.SEngine().Table("service").
		Join("INNER", "namespace", "service.namespace_id = namespace.id")
}

func getPipelineSession() *xorm.Session {
	return model.SEngine().Table("pipeline").
		Join("INNER", "service", "pipeline.service_id = service.id").
		Join("INNER", "namespace", "service.namespace_id = namespace.id")
}

func getImageSession() *xorm.Session {
	return model.SEngine().Table("pipeline_image").
		Join("INNER", "pipeline", "pipeline.id = pipeline_image.pipeline_id")
}

func getUpdateSession() *xorm.Session {
	return model.SEngine().Table("pipeline_update").
		Join("INNER", "pipeline", "pipeline_update.pipeline_id = pipeline.id").
		Join("INNER", "code_module", "pipeline_update.code_module_id = code_module.id").
		Join("INNER", "service", "pipeline.service_id = service.id")
}
