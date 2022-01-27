// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"time"

	"xorm.io/xorm"

	"nautilus/pkg/db"
)

func MEngine() *xorm.Engine {
	return db.MEngine
}

func SEngine() *xorm.Engine {
	return db.SEngine
}

type Namespace struct {
	ID       int64
	Name     string    `xorm:"varchar(32) notnull unique"`
	Creator  string    `xorm:"varchar(50) notnull"`
	CreateAt time.Time `xorm:"timestamp notnull created"`
}

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

type PipelineImage struct {
	ID         int64
	PipelineID int64     `xorm:"bigint notnull"`
	ImageURL   string    `xorm:"varchar(200)"`
	ImageTag   string    `xorm:"varchar(50)"`
	Status     int       `xorm:"int notnull"`
	CreateAt   time.Time `xorm:"timestamp notnull created"`
}

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
	PIWait    int = iota // 待构建
	PIProcess            // 构建中
	PISuccess            // 构建成功
	PIFailed             // 构建失败
)

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
