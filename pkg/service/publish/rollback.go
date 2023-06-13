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
	"nautilus/pkg/util/cm"
	"nautilus/pkg/util/k8s"
)

func NewRollback(pid int64, username string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	if cm.Ini(pipeline.Status, []int{model.PLRollbackSuccess, model.PLTerminate}) {
		return fmt.Errorf(config.ROL_CANNOT_EXECUTE)
	}

	svc, err := model.GetServiceInfo(pipeline.Service)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		rollbackGroup string // 回滚组恢复指定副本数
		destroyGroup  string // 销毁组缩成0
		namespace     = svc.Namespace
		service       = svc.Name
		replicas      = svc.Replicas
		serviceID     = svc.ID
	)

	log.Infof("before rollback, get online group(%s) deploy group(%s)", svc.OnlineGroup, svc.DeployGroup)

	// (1) 获取回滚组和销毁组
	if pipeline.Status == model.PLSuccess {
		// 发布成功时, 销毁当前在线组
		destroyGroup = svc.OnlineGroup
	} else {
		// 发布过程中, 销毁当前部署组
		destroyGroup = svc.DeployGroup
	}
	rollbackGroup = k8s.GetAnotherGroup(destroyGroup)
	log.Infof("get rollback group(%s) destroy group(%s)", rollbackGroup, destroyGroup)

	// (2) 占锁
	if err := model.UpdateStatus(pid, model.PLRollbacking); err != nil {
		return fmt.Errorf(config.DB_UPDATE_PIPELINE_ERROR, err)
	}

	if err := model.SetLock(serviceID, username); err != nil {
		return fmt.Errorf(config.DB_WRITE_LOCK_ERROR, pid, err)
	}

	// (3) 获取已发布的阶段
	phases, err := model.FindKindPhases(pid, model.PHASE_DEPLOY)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_PHASES_ERROR, err)
	}

	publishes := make([]string, 0)
	for _, obj := range phases {
		if obj.Status == model.PHProcess {
			return fmt.Errorf(config.ROL_PROCESS_NO_EXECUTE)
		}

		// 排除image、finish两个阶段
		if cm.In(obj.Name, []string{model.PHASE_IMAGE, model.PHASE_FINISH}) {
			continue
		}

		if cm.Ini(obj.Status, []int{model.PHSuccess, model.PHFailed}) {
			publishes = append(publishes, obj.Name)
		}
	}
	log.Infof("get published phase have: %s", publishes)

	// (4) 回滚组恢复指定副本、销毁组缩成0
	resource, err := k8s.New(namespace)
	if err != nil {
		return err
	}

	for _, phase := range publishes {
		if phase == model.PHASE_SANDBOX {
			replicas = 1
		}

		// 第一步: 回滚组恢复指定副本数
		rollbackDepName := k8s.GetDeploymentName(service, serviceID, phase, rollbackGroup)
		if err := resource.Scale(namespace, rollbackDepName, replicas); err != nil {
			log.Errorf("rollback deployment: %s replicas: %d error: %+v", rollbackDepName, replicas, err)
			return err
		}
		log.Infof("rollback deployment: %s replicas: %d success", rollbackDepName, replicas)

		// 第二步: 销毁组缩成0
		destroyDepName := k8s.GetDeploymentName(service, serviceID, phase, destroyGroup)
		if err := resource.Scale(namespace, destroyDepName, 0); err != nil {
			log.Errorf("destroy deployment: %s scale 0 error: %+v", destroyDepName, err)
			return err
		}
		log.Infof("destroy deployment: %s scale 0 success", destroyDepName)

		if err := model.CreatePhase(pid, model.PHASE_ROLLBACK, phase, model.PHSuccess); err != nil {
			return fmt.Errorf(config.ROL_RECORD_PHASE_ERROR, phase, err)
		}
	}

	// (5) 释放锁
	if err := model.UpdateStatus(pid, model.PLRollbackSuccess); err != nil {
		return fmt.Errorf(config.DB_UPDATE_PIPELINE_ERROR, err)
	}

	if err := model.SetLock(serviceID, ""); err != nil {
		return fmt.Errorf(config.DB_WRITE_LOCK_ERROR, pid, err)
	}
	log.Infof("rollback all phases: %+v success", publishes)

	// (6) 完成
	if err := model.UpdateGroup(pid, serviceID, rollbackGroup, destroyGroup, model.PLRollbackSuccess); err != nil {
		return fmt.Errorf(config.FSH_UPDATE_ONLINE_GROUP_ERROR, err)
	}
	return nil
}
