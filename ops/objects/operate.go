// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"fmt"

	"ferry/ops/db"
	"ferry/ops/log"
)

func CreatePipeline(name, summary, creator, rd, qa, pm, serviceName string, moduleInfoList []map[string]string) error {
	session := db.MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	service := new(db.Service)
	if has, err := session.Where("name=?", serviceName).Get(service); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("service query by name: %s is not exists", serviceName)
	}

	service.Lock = creator
	if _, err := session.ID(service.ID).Update(service); err != nil {
		return err
	}
	log.Infof("update service: %s lock: %s success", serviceName, creator)

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
	log.Infof("create pipeline success. get pipeline id: %d", pipeline.ID)

	for _, moduleInfo := range moduleInfoList {
		moduleName := moduleInfo["name"]
		deployBranch := moduleInfo["branch"]

		module := new(db.Module)
		if has, err := session.Where("name=? and service_id=?", moduleName, service.ID).Get(module); err != nil {
			return err
		} else if !has {
			return fmt.Errorf("query match is not exists")
		}

		pipelineUpdate := new(db.PipelineUpdate)
		pipelineUpdate.PipelineID = pipeline.ID
		pipelineUpdate.ModuleID = module.ID
		pipelineUpdate.DeployBranch = deployBranch
		if _, err := session.Insert(pipelineUpdate); err != nil {
			return err
		}
		log.Infof("create update info success. by module: %s branch: %s", moduleName, deployBranch)
	}
	return session.Commit()
}

func CreatePhase(pipelineID int64, name string, status int) error {
	phase := new(db.PipelinePhase)
	if has, err := db.MEngine.Where("pipeline_id=? and name=?", pipelineID, name).Get(phase); has {
		return nil
	} else if err != nil {
		return err
	}

	phase.Name = name
	phase.Status = status
	phase.PipelineID = pipelineID
	if _, err := db.MEngine.Insert(phase); err != nil {
		return err
	}
	return nil
}

func UpdatePhase(pipelineID int64, name string, status int, deployment string) error {
	phase := new(db.PipelinePhase)
	phase.Status = status
	phase.Deployment = deployment
	if affected, err := db.MEngine.Where("pipeline_id=? and name=?", pipelineID, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return fmt.Errorf("query match is not exists")
	}
	return nil
}

func UpdateGroup(pipelineID int64, serviceName, group string) error {
	session := db.MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	pipeline := new(db.Pipeline)
	pipeline.Status = db.PLSuccess
	if affected, err := session.ID(pipelineID).Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return fmt.Errorf("query match is not exists")
	}

	service := new(db.Service)
	service.OnlineGroup = group
	service.Lock = ""
	if affected, err := session.Where("name=?", serviceName).Cols("online_group", "lock").Update(service); err != nil {
		return err
	} else if affected == 0 {
		return fmt.Errorf("query match is not exists")
	}
	return session.Commit()
}
