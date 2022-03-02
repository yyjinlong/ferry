// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"fmt"
	"time"
)

type Service struct {
	ID            int64
	NamespaceID   int64     `xorm:"bigint notnull"`
	Name          string    `xorm:"varchar(32) notnull unique"`
	ImageAddr     string    `xorm:"varchar(500) notnull"`
	QuotaCpu      int       `xorm:"int"`
	QuotaMaxCpu   int       `xorm:"int"`
	QuotaMem      int       `xorm:"int"`
	QuotaMaxMem   int       `xorm:"int"`
	Replicas      int       `xorm:"int"`
	Volume        string    `xorm:"text"`
	ReserveTime   int       `xorm:"int"`
	Port          int       `xorm:"int"`
	ContainerPort int       `xorm:"int"`
	OnlineGroup   string    `xorm:"varchar(20) notnull"`
	DeployGroup   string    `xorm:"varchar(20) notnull"`
	MultiPhase    bool      `xorm:"bool"`
	Lock          string    `xorm:"varchar(100) notnull"`
	RD            string    `xorm:"varchar(50) notnull"`
	OP            string    `xorm:"varchar(50) notnull"`
	CreateAt      time.Time `xorm:"timestamp notnull created"`
	UpdateAt      time.Time `xorm:"timestamp notnull updated"`
}

type CodeModule struct {
	ID        int64
	ServiceID int64     `xorm:"bigint notnull"`
	Name      string    `xorm:"varchar(50) notnull"`
	Language  string    `xorm:"varchar(20) notnull"`
	ReposName string    `xorm:"varchar(10) notnull"`
	ReposAddr string    `xorm:"varchar(200)"`
	CreateAt  time.Time `xorm:"timestamp notnull created"`
	UpdateAt  time.Time `xorm:"timestamp notnull updated"`
}

func GetServiceInfo(name string) (*Service, error) {
	service := new(Service)
	if has, err := SEngine().Where("name = ?", name).Get(service); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return service, nil
}

func GetServiceByID(serviceID int64) (*Service, error) {
	service := new(Service)
	if has, err := SEngine().ID(serviceID).Get(service); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return service, nil
}

func GetCodeModuleInfo(module string) (*CodeModule, error) {
	codeModule := new(CodeModule)
	if has, err := SEngine().Where("name=?", module).Get(codeModule); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return codeModule, nil
}

func GetCodeModuleInfoByID(moduleID int64) (*CodeModule, error) {
	codeModule := new(CodeModule)
	if has, err := SEngine().ID(moduleID).Get(codeModule); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return codeModule, nil
}

func FindCodeModules(serviceID int64) ([]CodeModule, error) {
	moduleList := make([]CodeModule, 0)
	if err := SEngine().Where("service_id = ?", serviceID).Find(&moduleList); err != nil {
		return nil, err
	}
	return moduleList, nil
}

func UpdateTag(pipelineID int64, moduleName, codeTag string) error {
	session := MEngine().NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	codeModule := new(CodeModule)
	if has, err := session.Where("name = ?", moduleName).Get(codeModule); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("query module name: %s is not exists", moduleName)
	}

	pu := new(PipelineUpdate)
	pu.CodeTag = codeTag
	if affected, err := session.Where("pipeline_id=? and code_module_id=?",
		pipelineID, codeModule.ID).Cols("code_tag").Update(pu); err != nil {
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
