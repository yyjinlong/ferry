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

func NewFinish(pid int64) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	service, err := model.GetServiceInfo(pipeline.Service)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		namespace   = service.Name
		serviceID   = service.ID
		serviceName = service.Name
		onlineGroup = service.OnlineGroup
	)

	// 另一组缩成0
	resource, err := k8s.New(namespace)
	if err != nil {
		return err
	}

	for _, phase := range []string{model.PHASE_SANDBOX, model.PHASE_ONLINE} {
		oldDeployment := k8s.GetDeploymentName(serviceName, serviceID, phase, onlineGroup)
		if err := resource.Scale(namespace, oldDeployment, 0); err != nil {
			log.Errorf("old deployment: %s replicas scale 0 failed: %+v", oldDeployment, err)
			return err
		}
		log.Infof("old deployment: %s replicas scale 0 success", oldDeployment)
	}

	if err := model.CreatePhase(pid, model.PHASE_DEPLOY, model.PHASE_FINISH, model.PHSuccess); err != nil {
		return fmt.Errorf(config.FSH_CREATE_FINISH_PHASE_ERROR, err)
	}
	log.Infof("record finish phase for pid: %d success", pid)

	newOnlineGroup := service.DeployGroup
	newDeployGroup := k8s.GetDeployGroup(newOnlineGroup)
	log.Infof("get current online_group: %s deploy_group: %s", newOnlineGroup, newDeployGroup)

	if err := model.UpdateGroup(pid, service.ID, newOnlineGroup, newDeployGroup, model.PLSuccess); err != nil {
		return fmt.Errorf(config.FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	log.Infof("set current online group: %s deploy group: %s success", newOnlineGroup, newDeployGroup)
	return nil
}
