// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package rollback

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nautilus/pkg/config"
	"nautilus/pkg/k8s/exec"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func NewRollback() *Rollback {
	return &Rollback{}
}

type Rollback struct{}

func (ro *Rollback) Handle(pid int64, username string) error {
	// TODO: 建立websocket
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	if util.Ini(pipeline.Status, []int{model.PLRollbackSuccess, model.PLTerminate}) {
		return fmt.Errorf(config.ROL_CANNOT_EXECUTE)
	}

	// NOTE: 获取回滚组和销毁组
	service, err := model.GetServiceByID(pipeline.ServiceID)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		rollbackGroup string // 回滚组恢复指定副本数
		destroyGroup  string // 销毁组缩成0
	)

	if pipeline.Status == model.PLSuccess {
		// 发布成功时
		rollbackGroup = service.DeployGroup
		destroyGroup = service.OnlineGroup
	} else {
		// 发布过程中
		rollbackGroup = service.OnlineGroup
		destroyGroup = service.DeployGroup
	}
	log.Infof("get rollback_group: %s destroy_group: %s", rollbackGroup, destroyGroup)

	// NOTE: 占锁
	if err := model.UpdateStatus(pid, model.PLRollbacking); err != nil {
		return fmt.Errorf(config.DB_UPDATE_PIPELINE_ERROR, err)
	}

	if err := model.SetLock(pipeline.ServiceID, username); err != nil {
		return fmt.Errorf(config.DB_WRITE_LOCK_ERROR, pid, err)
	}

	// NOTE: 获取已发布的阶段
	phases, err := model.FindKindPhases(pid, model.PHASE_DEPLOY)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_PHASES_ERROR, err)
	}

	publishes := make([]string, 0)
	for _, obj := range phases {
		if obj.Status == model.PHProcess {
			return fmt.Errorf(config.ROL_PROCESS_NO_EXECUTE)
		}

		if util.Ini(obj.Status, []int{model.PHSuccess, model.PHFailed}) {
			publishes = append(publishes, obj.Name)
		}
	}
	log.Infof("get published phase have: %s", publishes)

	namespace, err := model.GetNamespace(service.NamespaceID)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_NAMESPACE_ERROR, err)
	}

	for _, phase := range publishes {
		replicas := service.Replicas
		if phase == model.PHASE_SANDBOX {
			replicas = 1
		}

		// NOTE: 回滚组恢复指定副本数
		rollbackDeployment := util.GetDeployment(service.Name, service.ID, phase, rollbackGroup)
		if err := ro.worker(namespace.Name, rollbackDeployment, replicas); err != nil {
			log.Errorf("rollback deployment: %s replicas: %d error: %+v", rollbackDeployment, replicas, err)
			return err
		}
		log.Infof("rollback deployment: %s replicas: %d finish", rollbackDeployment, replicas)

		// NOTE: 销毁组缩成0
		destroyDeployment := util.GetDeployment(service.Name, service.ID, phase, destroyGroup)
		if err := ro.worker(namespace.Name, destroyDeployment, 0); err != nil {
			log.Errorf("destroy deployment: %s scale 0 error: %+v", destroyDeployment, err)
			return err
		}

		if err := model.CreatePhase(pid, model.PHASE_ROLLBACK, phase, model.PHSuccess, ""); err != nil {
			return fmt.Errorf(config.ROL_RECORD_PHASE_ERROR, phase, err)
		}
		log.Infof("destroy deployment: %s scale 0 finish", destroyDeployment)
	}

	// NOTE: 释放锁
	if err := model.UpdateStatus(pid, model.PLRollbackSuccess); err != nil {
		return fmt.Errorf(config.DB_UPDATE_PIPELINE_ERROR, err)
	}

	if err := model.SetLock(pipeline.ServiceID, ""); err != nil {
		return fmt.Errorf(config.DB_WRITE_LOCK_ERROR, pid, err)
	}

	log.Infof("rollback phases: %+v success", publishes)
	return ro.finish(pid, service.ID, rollbackGroup, destroyGroup)
}

func (ro *Rollback) worker(namespace, deployment string, replicas int) error {
	dep := exec.NewDeployments(namespace, deployment)
	return dep.Scale(replicas)
}

func (ro *Rollback) finish(pid, serviceID int64, onlineGroup, deployGroup string) error {
	if err := model.UpdateGroup(pid, serviceID, onlineGroup, deployGroup, model.PLRollbackSuccess); err != nil {
		return fmt.Errorf(config.FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	log.Infof("set current online group: %s deploy group: %s for rollback success", onlineGroup, deployGroup)
	return nil
}
