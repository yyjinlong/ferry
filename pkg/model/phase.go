// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"time"

	"nautilus/pkg/util"
)

type PipelinePhase struct {
	ID              int64
	PipelineID      int64     `xorm:"bigint notnull"`
	Name            string    `xorm:"varchar(20)"`
	Kind            string    `xorm:"varchar(20)"`
	Status          int       `xorm:"int notnull"`
	Log             string    `xorm:"text"`
	Deployment      string    `xorm:"text"`
	ResourceVersion string    `xorm:"varchar(32)"`
	CreateAt        time.Time `xorm:"timestamp notnull created"`
	UpdateAt        time.Time `xorm:"timestamp notnull updated"`
}

// 状态定义
const (
	PHWait    int = iota // 待执行
	PHProcess            // 执行中
	PHSuccess            // 执行成功
	PHFailed             // 执行失败
)

// 阶段名称
const (
	PHASE_IMAGE   = "image"
	PHASE_SANDBOX = "sandbox"
	PHASE_ONLINE  = "online"
)

// 阶段类别
const (
	PHASE_DEPLOY   = "deploy"
	PHASE_ROLLBACK = "rollback"
)

var (
	PHASE_NAME_LIST = []string{PHASE_SANDBOX, PHASE_ONLINE}
)

// FindPhases 根据pipeline id返回对应的阶段
func FindPhases(pipelineID int64) ([]PipelinePhase, error) {
	ppList := make([]PipelinePhase, 0)
	if err := SEngine().Where("pipeline_id=?", pipelineID).Desc("id").Find(&ppList); err != nil {
		return nil, err
	}
	return ppList, nil
}

func CheckPhaseIsDeploy(pipelineID int64, kind, phase string) bool {
	ph := new(PipelinePhase)
	if has, err := SEngine().Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, phase).Get(ph); err != nil {
		return false
	} else if !has {
		return false
	}
	return true
}

func GetPhaseInfo(pipelineID int64, kind, phase string) (*PipelinePhase, error) {
	phaseObj := new(PipelinePhase)
	if has, err := SEngine().Where("pipeline_id = ? and kind = ? and name = ?", pipelineID, kind, phase).Get(phaseObj); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return phaseObj, nil
}

func CreatePhase(pipelineID int64, kind, name string, status int, deployment string) error {
	phase := new(PipelinePhase)
	if has, err := MEngine().Where("pipeline_id=? and kind =? and name=?", pipelineID, kind, name).Get(phase); has {
		return nil
	} else if err != nil {
		return err
	}

	phase.Name = name
	phase.Kind = kind
	phase.Status = status
	phase.PipelineID = pipelineID
	phase.Deployment = deployment
	if _, err := MEngine().Insert(phase); err != nil {
		return err
	}
	return nil
}

func UpdatePhase(pipelineID int64, kind, name string, status int) error {
	phase := new(PipelinePhase)
	phase.Status = status
	if affected, err := MEngine().Cols("status").Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func UpdatePhaseV2(pipelineID int64, kind, name string, status int, version string) error {
	phase := new(PipelinePhase)
	phase.Status = status
	phase.ResourceVersion = version
	if affected, err := MEngine().Cols("status", "resource_version").Where(
		"pipeline_id=? and kind=? and name=?", pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func RealtimeLog(pipelineID int64, kind, name, msg string) error {
	phase := new(PipelinePhase)
	if has, err := MEngine().Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Get(phase); err != nil {
		return err
	} else if !has {
		return NotFound
	}

	if util.Ini(phase.Status, []int{PHSuccess, PHFailed}) {
		return nil
	}

	phase.Log = phase.Log + "\n" + msg
	if affected, err := MEngine().Cols("log").Where("pipeline_id=? and kind=? and name=?",
		pipelineID, kind, name).Update(phase); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}
