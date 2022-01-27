// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package objects

import (
	"fmt"

	"nautilus/internal/model"
	"nautilus/pkg/g"
	"nautilus/pkg/log"
)

func CreatePipeline(name, summary, creator, rd, qa, pm, serviceName string, moduleInfoList []map[string]string) error {
	session := model.MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	service := new(model.Service)
	if has, err := session.Where("name=?", serviceName).Get(service); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("service query by name: %s is not exists", serviceName)
	}

	pipeline := new(model.Pipeline)
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

		codeModule := new(model.CodeModule)
		if has, err := session.Where("name=? and service_id=?", moduleName, service.ID).Get(codeModule); err != nil {
			return err
		} else if !has {
			return NotFound
		}

		pipelineUpdate := new(model.PipelineUpdate)
		pipelineUpdate.PipelineID = pipeline.ID
		pipelineUpdate.CodeModuleID = codeModule.ID
		pipelineUpdate.DeployBranch = deployBranch
		if _, err := session.Insert(pipelineUpdate); err != nil {
			return err
		}
		log.Infof("create update info success. by code_module: %s branch: %s", moduleName, deployBranch)
	}
	return session.Commit()
}

func SetLock(serviceID int64, lock string) error {
	service := new(model.Service)
	service.Lock = lock
	if _, err := model.MEngine().ID(serviceID).Update(service); err != nil {
		return err
	}
	return nil
}

func CreateImage(pipelineID int64) error {
	image := new(model.PipelineImage)
	image.PipelineID = pipelineID
	image.Status = model.PIProcess
	if _, err := model.MEngine().Insert(image); err != nil {
		return err
	}
	return nil
}

func UpdateImage(pipelineID int64, imageURL, imageTag string) error {
	image := new(model.PipelineImage)
	image.PipelineID = pipelineID
	image.ImageURL = imageURL
	image.ImageTag = imageTag
	if _, err := model.MEngine().Where("pipeline_id=?", pipelineID).Update(image); err != nil {
		return err
	}
	return nil
}

func CreatePhase(pipelineID int64, kind, name string, status int, deployment string) error {
	phase := new(model.PipelinePhase)
	if has, err := model.MEngine().Where("pipeline_id=? and kind =? and name=?", pipelineID, kind, name).Get(phase); has {
		return nil
	} else if err != nil {
		return err
	}

	phase.Name = name
	phase.Kind = kind
	phase.Status = status
	phase.PipelineID = pipelineID
	phase.Deployment = deployment
	if _, err := model.MEngine().Insert(phase); err != nil {
		return err
	}
	return nil
}

func UpdatePhase(pipelineID int64, kind, name string, status int) error {
	phase := new(model.PipelinePhase)
	phase.Status = status
	if affected, err := model.MEngine().Cols("status").Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdatePhaseV2(pipelineID int64, kind, name string, status int, version string) error {
	phase := new(model.PipelinePhase)
	phase.Status = status
	phase.ResourceVersion = version
	if affected, err := model.MEngine().Cols("status", "resource_version").Where(
		"pipeline_id=? and kind=? and name=?", pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdateGroup(pipelineID int64, serviceName, onlineGroup, deployGroup string) error {
	session := model.MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	pipeline := new(model.Pipeline)
	pipeline.Status = model.PLSuccess
	if affected, err := session.ID(pipelineID).Cols("status").Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}

	service := new(model.Service)
	service.OnlineGroup = onlineGroup
	service.DeployGroup = deployGroup
	service.Lock = ""
	if affected, err := session.Where("name=?", serviceName).Cols("online_group", "deploy_group", "lock").Update(service); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return session.Commit()
}

func UpdateTag(pipelineID int64, moduleName, codeTag string) error {
	session := model.MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	codeModule := new(model.CodeModule)
	if has, err := session.Where("name = ?", moduleName).Get(codeModule); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("query module name: %s is not exists", moduleName)
	}

	pu := new(model.PipelineUpdate)
	pu.CodeTag = codeTag
	if affected, err := session.Where("pipeline_id=? and code_module_id=?",
		pipelineID, codeModule.ID).Cols("code_tag").Update(pu); err != nil {
	} else if affected == 0 {
		return NotFound
	}

	pipeline := new(model.Pipeline)
	pipeline.Status = model.PLProcess
	if affected, err := session.ID(pipelineID).Cols("status").Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}

	return session.Commit()
}

func RealtimeLog(pipelineID int64, kind, name, msg string) error {
	phase := new(model.PipelinePhase)
	if has, err := model.MEngine().Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Get(phase); err != nil {
		return err
	} else if !has {
		return NotFound
	}

	if g.Ini(phase.Status, []int{model.PHSuccess, model.PHFailed}) {
		return nil
	}

	phase.Log = phase.Log + "\n" + msg
	if affected, err := model.MEngine().Cols("log").Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}
