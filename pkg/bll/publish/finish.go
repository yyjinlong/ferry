// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"errors"
	"fmt"

	"nautilus/internal/objects"
	"nautilus/pkg/base"
	"nautilus/pkg/log"
)

type Finish struct{}

func (f *Finish) Handle(r *base.Request) (interface{}, error) {
	type params struct {
		ID int64 `form:"pipeline_id" binding:"required"`
	}

	var data params
	if err := r.ShouldBind(&data); err != nil {
		return nil, err
	}

	pid := data.ID
	log.InitFields(log.Fields{"logid": r.TraceID, "pipeline_id": pid})

	pipeline, err := objects.GetPipelineInfo(pid)
	if errors.Is(err, objects.NotFound) {
		return nil, fmt.Errorf(DB_PIPELINE_NOT_FOUND, pid)
	} else if err != nil {
		return nil, fmt.Errorf(DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	onlineGroup := pipeline.Service.DeployGroup
	deployGroup := f.getDeployGroup(onlineGroup)
	log.Infof("get current online group: %s deploy group: %s", onlineGroup, deployGroup)

	if err := objects.UpdateGroup(pid, pipeline.Service.Name, onlineGroup, deployGroup); err != nil {
		log.Errorf("update finish info error: %s", err)
		return nil, fmt.Errorf(FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	log.Infof("set current online group: %s deploy group: %s success.", onlineGroup, deployGroup)
	return "", nil
}

func (f *Finish) getDeployGroup(group string) string {
	if group == objects.BLUE {
		return objects.GREEN
	}
	return objects.BLUE
}
