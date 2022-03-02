// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	"github.com/yyjinlong/golib/log"

	"nautilus/pkg/cfg"
	"nautilus/pkg/cm"
	"nautilus/pkg/model"
)

func NewFinish() *Finish {
	return &Finish{}
}

type Finish struct{}

func (f *Finish) Handle(pid int64) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(cfg.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	serviceObj, err := model.GetServiceByID(pipeline.ServiceID)
	if err != nil {
		return fmt.Errorf(cfg.DB_SERVICE_QUERY_ERROR, err)
	}

	serviceName := serviceObj.Name
	onlineGroup := serviceObj.DeployGroup
	deployGroup := cm.GetDeployGroup(onlineGroup)
	log.Infof("get current online group: %s deploy group: %s", onlineGroup, deployGroup)

	if err := model.UpdateGroup(pid, serviceName, onlineGroup, deployGroup); err != nil {
		return fmt.Errorf(cfg.FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	log.Infof("set current online group: %s deploy group: %s success.", onlineGroup, deployGroup)
	return nil
}
