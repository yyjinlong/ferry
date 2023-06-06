// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"time"

	"xorm.io/xorm"
)

// PipelineImage 每次上线存储全量模块的镜像信息
type PipelineImage struct {
	ID         int64
	PipelineID int64     `xorm:"bigint notnull"`
	Service    string    `xorm:"varchar(32) notnull"` // 服务
	CodeModule string    `xorm:"varchar(50) notnull"` // 代码模块
	ImageURL   string    `xorm:"varchar(200)"`        // 对应代码模块镜像地址
	ImageTag   string    `xorm:"varchar(50)"`         // 对应代码模块镜像tag
	Status     int       `xorm:"int notnull"`
	CreateAt   time.Time `xorm:"timestamp notnull created"`
}

type ImageUnionQuery struct {
	Pipeline      `xorm:"extends"`
	PipelineImage `xorm:"extends"`
}

const (
	PIWait    int = iota // 待构建
	PIProcess            // 构建中
	PISuccess            // 构建成功
	PIFailed             // 构建失败
)

func ImageSession() *xorm.Session {
	return SEngine.Table("pipeline").Alias("p").
		Join("INNER", []string{"pipeline_image", "pi"}, "p.id = pi.pipeline_id")
}

func CreateImage(pipelineID int64, codeModule string) error {
	image := new(PipelineImage)
	image.PipelineID = pipelineID
	image.CodeModule = codeModule
	image.Status = PIProcess
	if _, err := MEngine.Insert(image); err != nil {
		return err
	}
	return nil
}

func UpdateImage(pipelineID int64, codeModule, imageURL, imageTag string) error {
	image := new(PipelineImage)
	image.PipelineID = pipelineID
	image.CodeModule = codeModule
	image.ImageURL = imageURL
	image.ImageTag = imageTag
	if _, err := MEngine.Where("pipeline_id=?", pipelineID).Update(image); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdatePipelineImage(pipelineID int64, service, codeModule, imageURL, imageTag string) error {
	table := new(PipelineImage)
	if has, err := SEngine.Where("pipeline_id=? and service=? and code_module=?", pipelineID, service, codeModule).Get(table); err != nil {
		return err
	} else if !has {
		table.PipelineID = pipelineID
		table.Service = service
		table.CodeModule = codeModule
		table.ImageURL = imageURL
		table.ImageTag = imageTag
		if _, err := MEngine.Insert(table); err != nil {
			return err
		}
		return nil
	}

	table.ImageURL = imageURL
	table.ImageTag = imageTag
	if affected, err := MEngine.Cols("iamge_url", "image_tag", "update_at").
		Where("pipeline_id=? and service=? and code_module=?", pipelineID, service, codeModule).Update(table); err != nil {
		return err
	} else if affected == 0 {
		return NotFound
	}
	return nil
}

func FindImages(pipelineID int64) ([]PipelineImage, error) {
	images := make([]PipelineImage, 0)
	if err := SEngine.Where("pipeline_id = ?", pipelineID).Find(&images); err != nil {
		return nil, err
	}
	return images, nil
}

func QueryLatestSuccessModuleImage(service, codeModule string) (*ImageUnionQuery, error) {
	image := new(ImageUnionQuery)
	if has, err := ImageSession().Where("p.status=? and p.service=? and pi.code_module=?",
		PLSuccess, service, codeModule).Get(image); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return image, nil
}
