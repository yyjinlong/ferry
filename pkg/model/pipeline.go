// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Pipeline struct {
	ID       int64
	Service  string    `xorm:"varchar(32) notnull"`
	Name     string    `xorm:"varchar(100) notnull"`
	Summary  string    `xorm:"text notnull"`
	Creator  string    `xorm:"varchar(50) notnull"`
	RD       string    `xorm:"varchar(500) notnull"`
	QA       string    `xorm:"varchar(200)"`
	PM       string    `xorm:"varchar(500) notnull"`
	Status   int       `xorm:"int notnull"`
	CreateAt time.Time `xorm:"timestamp notnull created"`
	UpdateAt time.Time `xorm:"timestamp notnull updated"`
}

type PipelineUpdate struct {
	ID           int64
	PipelineID   int64     `xorm:"bigint notnull"`
	CodeModule   string    `xorm:"varchar(50) notnull"`
	DeployBranch string    `xorm:"varchar(20)"`
	CodeTag      string    `xorm:"varchar(50)"`
	CodePkg      string    `xorm:"varchar(100)"`
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
	if has, err := SEngine.ID(pipelineID).Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// GetServicePipeline 根据服务返回最近一次的上线信息
func GetServicePipeline(service string) (*Pipeline, error) {
	pipeline := new(Pipeline)
	if has, err := SEngine.Where("service = ? and status != 0", service).Desc("id").Limit(1).Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// GetServiceLastSuccessPipeline 根据服务返回最近一次成功的上线信息
func GetServiceLastSuccessPipeline(service string) (*Pipeline, error) {
	pipeline := new(Pipeline)
	if has, err := SEngine.Where("service = ? AND pipeline.status = ?", service, PLSuccess).
		Desc("id").Get(pipeline); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pipeline, nil
}

// FindPipelineInfo 根据service返回pipeline相关信息
func FindPipelineInfo(service string) ([]Pipeline, error) {
	pList := make([]Pipeline, 0)
	if err := SEngine.Where("service = ? AND pipeline.status = ?", service, PLSuccess).
		Desc("id").Find(&pList); err != nil {
		return nil, err
	}
	return pList, nil
}

func FindUpdateInfo(pipelineID int64) ([]PipelineUpdate, error) {
	uqList := make([]PipelineUpdate, 0)
	if err := SEngine.Where("pipeline_id = ?", pipelineID).Find(&uqList); err != nil {
		return nil, err
	}
	return uqList, nil
}

func CreatePipeline(name, summary, creator, rd, qa, pm, serviceName string, moduleInfoList []map[string]string) error {
	session := MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	pipeline := new(Pipeline)
	pipeline.Name = name
	pipeline.Summary = summary
	pipeline.Creator = creator
	pipeline.RD = rd
	pipeline.QA = qa
	pipeline.PM = pm
	pipeline.Service = serviceName
	if _, err := session.Insert(pipeline); err != nil {
		return err
	}
	log.Infof("create pipeline success. get pipeline id: %d", pipeline.ID)

	for _, moduleInfo := range moduleInfoList {
		moduleName := moduleInfo["name"]
		deployBranch := moduleInfo["branch"]

		pipelineUpdate := new(PipelineUpdate)
		pipelineUpdate.PipelineID = pipeline.ID
		pipelineUpdate.CodeModule = moduleName
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
	if _, err := MEngine.Cols("lock").ID(serviceID).Update(service); err != nil {
		return err
	}
	return nil
}

func UpdateTag(pipelineID int64, moduleName, codeTag string) error {
	session := MEngine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	pu := new(PipelineUpdate)
	pu.CodeTag = codeTag
	if affected, err := MEngine.Where("pipeline_id=? and code_module=?",
		pipelineID, moduleName).Cols("code_tag").Update(pu); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}

	pipeline := new(Pipeline)
	pipeline.Status = PLProcess
	if affected, err := session.ID(pipelineID).Cols("status").Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}

	return session.Commit()
}

func UpdatePkg(pipelineID int64, moduleName, codePkg string) error {
	pu := new(PipelineUpdate)
	pu.CodePkg = codePkg
	if affected, err := MEngine.Where("pipeline_id=? and code_module=?",
		pipelineID, moduleName).Cols("code_pkg").Update(pu); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdateStatus(pipelineID int64, status int) error {
	pipeline := new(Pipeline)
	pipeline.Status = status
	if affected, err := MEngine.Cols("status", "update_at").ID(pipelineID).Update(pipeline); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdateGroup(pipelineID, serviceID int64, onlineGroup, deployGroup string, status int) error {
	session := MEngine.NewSession()
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
