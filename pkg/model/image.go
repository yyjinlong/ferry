// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"time"
)

type PipelineImage struct {
	ID         int64
	PipelineID int64     `xorm:"bigint notnull"`
	ImageURL   string    `xorm:"varchar(200)"`
	ImageTag   string    `xorm:"varchar(50)"`
	Status     int       `xorm:"int notnull"`
	CreateAt   time.Time `xorm:"timestamp notnull created"`
}

const (
	PIWait    int = iota // 待构建
	PIProcess            // 构建中
	PISuccess            // 构建成功
	PIFailed             // 构建失败
)

func GetImagInfo(pipelineID int64) (*PipelineImage, error) {
	pi := new(PipelineImage)
	if has, err := SEngine.Where("pipeline_id=?", pipelineID).Get(pi); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return pi, nil
}

// FindImageInfo 根据pipeline id返回本次构建的镜像信息
func FindImageInfo(pipelineID int64) (map[string]string, error) {
	pi := new(PipelineImage)
	if has, err := SEngine.Where("pipeline_id = ?", pipelineID).Get(pi); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}

	imageInfo := map[string]string{
		"image_url": pi.ImageURL,
		"image_tag": pi.ImageTag,
	}
	return imageInfo, nil
}

func CreateImage(pipelineID int64) error {
	image := new(PipelineImage)
	image.PipelineID = pipelineID
	image.Status = PIProcess
	if _, err := MEngine.Insert(image); err != nil {
		return err
	}
	return nil
}

func UpdateImage(pipelineID int64, imageURL, imageTag string) error {
	image := new(PipelineImage)
	image.PipelineID = pipelineID
	image.ImageURL = imageURL
	image.ImageTag = imageTag
	if _, err := MEngine.Where("pipeline_id=?", pipelineID).Update(image); err != nil {
		return err
	}
	return nil
}
