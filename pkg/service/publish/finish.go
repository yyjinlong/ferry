// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/k8s"
)

func NewFinish() *Finish {
	return &Finish{}
}

type Finish struct{}

func (f *Finish) Handle(pid int64) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	service, err := model.GetServiceByID(pipeline.ServiceID)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		serviceID   = service.ID
		serviceName = service.Name
		onlineGroup = service.DeployGroup
		deployGroup = k8s.GetDeployGroup(onlineGroup)
	)
	log.Infof("get current online_group: %s deploy_group: %s", onlineGroup, deployGroup)

	// 另一组缩成0
	for _, phase := range []string{model.PHASE_SANDBOX, model.PHASE_ONLINE} {
		oldDeployment := k8s.GetDeploymentName(serviceName, serviceID, phase, onlineGroup)
		fmt.Println("----缩成0的deployment: ", oldDeployment)
	}

	if err := model.UpdateGroup(pid, service.ID, onlineGroup, deployGroup, model.PLSuccess); err != nil {
		return fmt.Errorf(config.FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	log.Infof("set current online group: %s deploy group: %s success", onlineGroup, deployGroup)
	return nil
}
