// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/log"
)

type BuildTag struct {
	pid         int64
	serviceName string
	module      string
	tag         string
}

func (bt *BuildTag) Handle(c *gin.Context, r *base.MyRequest) (interface{}, error) {
	type params struct {
		ID      int64  `form:"pipeline_id" binding:"required"`
		Service string `form:"service" binding:"required"`
		Module  string `form:"module" binding:"required"`
		Tag     string `form:"tag" binding:"required"`
	}

	var data params
	if err := c.ShouldBind(&data); err != nil {
		return "", err
	}

	bt.pid = data.ID
	bt.serviceName = data.Service
	bt.module = data.Module
	bt.tag = data.Tag
	pidStr := strconv.FormatInt(bt.pid, 10)
	log.InitFields(log.Fields{"logid": r.RequestID, "pipeline_id": bt.pid})

	serviceObj, err := objects.GetServiceInfo(bt.serviceName)
	if err != nil {
		return "", fmt.Errorf(DB_QUERY_SERVICE_ERROR, bt.serviceName, err)
	}

	if serviceObj.Lock != "" && serviceObj.Lock != pidStr {
		return "", fmt.Errorf(TAG_OPERATE_FORBIDDEN, pidStr)
	}

	if err := objects.SetLock(serviceObj.Service.ID, pidStr); err != nil {
		return "", fmt.Errorf(TAG_WRITE_LOCK_ERROR, pidStr, err)
	}

	if err := bt.maketag(); err != nil {
		return "", fmt.Errorf(TAG_EXECUTE_SCRIPT_ERROR, err)
	}

	if err := objects.UpdateTag(bt.pid, bt.module, bt.tag); err != nil {
		return "", fmt.Errorf(TAG_UPDATE_DB_ERROR, err)
	}
	log.Infof("module: %s update tag: %s success", bt.module, bt.tag)
	return "", nil
}

func (bt *BuildTag) maketag() error {

	return nil
}
