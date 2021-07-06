// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"fmt"

	"xorm.io/xorm"

	"ferry/ops/db"
)

func CreatePipeline(name, summary, creator, rd, qa, pm, serviceName string, moduleInfoList []map[string]string) error {
	session := db.MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	service := new(db.Service)
	if has, err := session.Where("name=?").Get(service); !has {
		return err
	}

	pipeline := new(db.Pipeline)
	pipeline.Name = name
	pipeline.Summary = summary
	pipeline.Creator = creator
	pipeline.RD = rd
	pipeline.QA = qa
	pipeline.PM = pm
	pipeline.ServiceID = service.ID
	if _, err := session.Insert(pipeline); err != nil {
		return err
	}

	fmt.Println("----获取插入的pipeline id为: ", pipeline.ID)

	for _, moduleInfo := range moduleInfoList {
		moduleName := moduleInfo["name"]
		deployBranch := moduleInfo["branch"]

		module := new(db.Module)
		if has, err := session.Where("name=? and service_id=?", moduleName, service.ID).Get(module); !has {
			return err
		}

		pipelineUpdate := new(db.PipelineUpdate)
		pipelineUpdate.PipelineID = pipeline.ID
		pipelineUpdate.ModuleID = module.ID
		pipelineUpdate.DeployBranch = deployBranch
		if _, err := session.Insert(pipelineUpdate); err != nil {
			return err
		}
	}

	return session.Commit()
}

func CreatePhase(session *xorm.Session, pipelineID int64, name string, status int) error {
	phase := new(db.PipelinePhase)
	if has, err := session.Where("pipeline_id=? and name=?", pipelineID, name).Get(phase); has {
		return nil
	} else if err != nil {
		return err
	}

	phase.Name = name
	phase.Status = status
	phase.PipelineID = pipelineID
	if _, err := session.Insert(phase); err != nil {
		return err
	}
	return nil
}

func UpdatePhase(session *xorm.Session, pipelineID int64, name string, status int, deployment string) error {
	phase := new(db.PipelinePhase)
	phase.Status = status
	phase.Deployment = deployment
	if affected, err := session.Where("pipeline_id=? and name=?", pipelineID, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return fmt.Errorf("query match is not exists")
	}
	return nil
}
