// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

// 关联查询定义

type PipelineQuery struct {
	Pipeline  `xorm:"extends"`
	Service   `xorm:"extends"`
	Namespace `xorm:"extends"`
}

type ImageQuery struct {
	PipelineImage `xorm:"extends"`
	Pipeline      `xorm:"extends"`
}

type UpdateQuery struct {
	PipelineUpdate `xorm:"extends"`
	Pipeline       `xorm:"extends"`
	CodeModule     `xorm:"extends"`
	Service        `xorm:"extends"`
}

type ServiceQuery struct {
	Service   `xorm:"extends"`
	Namespace `xorm:"extends"`
}
