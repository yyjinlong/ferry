// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"fmt"
	"time"

	"nautilus/golib/log"
)

type Pipeline struct {
	ID        int64
	ServiceID int64     `xorm:"bigint notnull"`
	Name      string    `xorm:"varchar(100) notnull"`
	Summary   string    `xorm:"text notnull"`
	Creator   string    `xorm:"varchar(50) notnull"`
	RD        string    `xorm:"varchar(500) notnull"`
	QA        string    `xorm:"varchar(200)"`
	PM        string    `xorm:"varchar(500) notnull"`
	Status    int       `xorm:"int notnull"`
	CreateAt  time.Time `xorm:"timestamp notnull created"`
	UpdateAt  time.Time `xorm:"timestamp notnull updated"`
}

type PipelineUpdate struct {
	ID           int64
	PipelineID   int64     `xorm:"bigint notnull"`
	CodeModuleID int64     `xorm:"bigint notnull"`
	DeployBranch string    `xorm:"varchar(20)"`
	CodeTag      string    `xorm:"varchar(50)"`
	CreateAt     time.Time `xorm:"timestamp notnull created"`
}

const (
	PLWait            int = iota // 待上线
	PLProcess                    // 上线中
	PLSuccess                    // 上线成功
	PLFailed                     // 上线失败
	PLRollbacking                // 回滚中
	PLRollbackSuccess            // 回滚成功
	PLRollbackFailed             // 回滚失败
	PLTerminate                  // 流程终止
)

func GetPipeline(pipelineID int64) (*Pipeline, error) {
	pipeline := new(Pipeline)
	if has, err := SEngine().ID(pipelineID).Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// GetServicePipeline 根据服务id返回最近一次的上线信息
func GetServicePipeline(serviceID int64) (*Pipeline, error) {
	pipeline := new(Pipeline)
	if has, err := SEngine().Where("service_id = ?", serviceID).Desc("id").Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// FindPipelineInfo 根据service返回pipeline相关信息
func FindPipelineInfo(serviceID int64) ([]Pipeline, error) {
	pList := make([]Pipeline, 0)
	if err := SEngine().Where("service_id = ? AND pipeline.status = ?", serviceID, PLSuccess).
		Find(&pList); err != nil {
		return nil, err
	}
	return pList, nil
}

func FindUpdateInfo(pipelineID int64) ([]PipelineUpdate, error) {
	uqList := make([]PipelineUpdate, 0)
	if err := SEngine().Where("pipeline_id = ?", pipelineID).Find(&uqList); err != nil {
		return nil, err
	}
	return uqList, nil
}

func CreatePipeline(name, summary, creator, rd, qa, pm, serviceName string, moduleInfoList []map[string]string) error {
	session := MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	service := new(Service)
	if has, err := session.Where("name=?", serviceName).Get(service); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("service query by name: %s is not exists", serviceName)
	}

	pipeline := new(Pipeline)
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

		codeModule := new(CodeModule)
		if has, err := session.Where("name=? and service_id=?", moduleName, service.ID).Get(codeModule); err != nil {
			return err
		} else if !has {
			return NotFound
		}

		pipelineUpdate := new(PipelineUpdate)
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
	service := new(Service)
	service.Lock = lock
	if _, err := MEngine().Cols("lock").ID(serviceID).Update(service); err != nil {
		return err
	}
	return nil
}

func UpdateStatus(pipelineID int64, status int) error {
	pipeline := new(Pipeline)
	pipeline.Status = status
	if affected, err := MEngine().Cols("status", "update_at").ID(pipelineID).Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdateGroup(pipelineID, serviceID int64, onlineGroup, deployGroup string, status int) error {
	session := MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	pipeline := new(Pipeline)
	pipeline.Status = status
	if affected, err := session.ID(pipelineID).Cols("status", "update_at").Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}

	service := new(Service)
	service.OnlineGroup = onlineGroup
	service.DeployGroup = deployGroup
	service.Lock = ""
	if affected, err := session.ID(serviceID).Cols("online_group", "deploy_group", "lock").Update(service); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return session.Commit()
}
