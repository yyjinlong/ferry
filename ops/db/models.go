// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package db

import (
	"time"
)

type Namespace struct {
	ID       int64
	Name     string    `xorm:"varchar(32) notnull unique"`
	Creator  string    `xorm:"varchar(50) notnull"`
	CreateAt time.Time `xorm:"timestamp notnull created"`
}

type Service struct {
	ID          int64
	NamespaceID int64     `xorm:"bigint notnull"`
	Name        string    `xorm:"varchar(32) notnull unique"`
	Replicas    int       `xorm:"int"`
	Container   string    `xorm:"text"`
	Volume      string    `xorm:"text"`
	OnlineGroup string    `xorm:"varchar(20) notnull"`
	Lock        string    `xorm:"varchar(100) notnull"`
	ReserveTime int       `xorm:"int"`
	DeployPath  string    `xorm:"varchar(100)"`
	MultiPhase  bool      `xorm:"bool"`
	RD          string    `xorm:"varchar(50) notnull"`
	OP          string    `xorm:"varchar(50) notnull"`
	CreateAt    time.Time `xorm:"timestamp notnull created"`
	UpdateAt    time.Time `xorm:"timestamp notnull updated"`
}

type Module struct {
	ID        int64
	ServiceID int64     `xorm:"bigint notnull"`
	Name      string    `xorm:"varchar(50) notnull"`
	Language  string    `xorm:"varchar(20) notnull"`
	ReposName string    `xorm:"varchar(10) notnull"`
	ReposAddr string    `xorm:"varchar(200)"`
	CreateAt  time.Time `xorm:"timestamp notnull created"`
	UpdateAt  time.Time `xorm:"timestamp notnull updated"`
}

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
	ModuleID     int64     `xorm:"bigint notnull"`
	DeployBranch string    `xorm:"varchar(20)"`
	CodeTag      string    `xorm:"varchar(50)"`
	CreateAt     time.Time `xorm:"timestamp notnull created"`
}

type PipelineImage struct {
	ID         int64
	PipelineID int64     `xorm:"bigint notnull"`
	ModuleID   int64     `xorm:"bigint notnull"`
	ImageURL   string    `xorm:"varchar(200)"`
	ImageTag   string    `xorm:"varchar(50)"`
	CreateAt   time.Time `xorm:"timestamp notnull created"`
}

type PipelinePhase struct {
	ID         int64
	PipelineID int64     `xorm:"bigint notnull"`
	Name       string    `xorm:"varchar(20)"`
	Status     int       `xorm:"int notnull"`
	Log        string    `xorm:"text"`
	Deployment string    `xorm:"text"`
	CreateAt   time.Time `xorm:"timestamp notnull created"`
	UpdateAt   time.Time `xorm:"timestamp notnull updated"`
}

// 状态定义

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

const (
	PHWait    int = iota // 待执行
	PHProcess            // 执行中
	PHSuccess            // 执行成功
	PHFailed             // 执行失败
)

// 关联查询定义

type PipelineQuery struct {
	Pipeline  `xorm:"extends"`
	Service   `xorm:"extends"`
	Namespace `xorm:"extends"`
}

type ImageQuery struct {
	PipelineImage `xorm:"extends"`
	Pipeline      `xorm:"extends"`
	Module        `xorm:"extends"`
}
