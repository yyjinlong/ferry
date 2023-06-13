// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
)

type Deployment interface {
	HandleDeployment(obj interface{}, mode, cluster string) error
}

type DeploymentResource struct {
	clientset *kubernetes.Clientset
}

func NewDeploymentResource(clientset *kubernetes.Clientset) *DeploymentResource {
	return &DeploymentResource{
		clientset: clientset,
	}
}

func (r *DeploymentResource) HandleDeployment(obj interface{}, mode, cluster string) error {
	var (
		data              = obj.(*appsv1.Deployment)
		name              = data.ObjectMeta.Name
		namespace         = data.ObjectMeta.Namespace
		replicas          = *data.Spec.Replicas
		updatedReplicas   = data.Status.UpdatedReplicas
		readyReplicas     = data.Status.ReadyReplicas
		availableReplicas = data.Status.AvailableReplicas
	)

	// 检测是否是业务的deployment
	if !r.filter(name) {
		return nil
	}

	// 另一组缩成0, 不进行处理
	if replicas == 0 {
		return nil
	}

	// 检查deployment是否ready
	if !(data.ObjectMeta.Generation == data.Status.ObservedGeneration &&
		replicas == updatedReplicas &&
		replicas == availableReplicas &&
		replicas == readyReplicas) {
		return nil
	}

	serviceName, phase, group := r.parseInfo(name)
	log.Infof("[deployment] check: %s ready with group: %s replicas: %d", name, group, replicas)

	pipeline, err := model.GetServicePipeline(serviceName)
	if errors.Is(err, model.NotFound) {
		return nil
	} else if err != nil {
		log.Errorf("[deployment] query pipeline by service: %s error: %s", serviceName, err)
		return err
	}
	pipelineID := pipeline.ID

	if r.checkPipelineFinish(pipeline.Status) {
		return nil
	}

	if !r.checkSameNamespace(namespace, serviceName) {
		return nil
	}

	kind := model.PHASE_DEPLOY
	if pipeline.Status == model.PLRollbacking {
		kind = model.PHASE_ROLLBACK
	}
	log.Infof("[deployment] get pipeline: %d deploy kind is: %s", pipelineID, kind)
	if r.checkPhaseFinish(pipelineID, kind, phase) {
		return nil
	}

	// deployment ready更新阶段完成
	if err := model.UpdatePhaseV2(pipelineID, kind, phase, model.PHSuccess); err != nil {
		log.Errorf("[deployment] update pipeline: %d to db on phase: %s failed: %s", pipelineID, phase, err)
		return err
	}
	log.Infof("[deployment] update pipeline: %d to db on phase: %s success", pipelineID, phase)
	return nil
}

func (r *DeploymentResource) filter(name string) bool {
	re := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if re == nil {
		return false
	}
	result := re.FindAllStringSubmatch(name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (r *DeploymentResource) parseInfo(name string) (string, string, string) {
	re := regexp.MustCompile(`-\d+-`)

	// 获取服务名、阶段、蓝绿组
	matchList := re.Split(name, -1)
	serviceName := matchList[0]

	afterList := strings.Split(matchList[1], "-")
	phase := afterList[0]
	group := afterList[1]
	return serviceName, phase, group
}

func (r *DeploymentResource) checkPipelineFinish(status int) bool {
	if cm.Ini(status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Info("[deployment] check deploy is finished")
		return true
	}
	return false
}

func (r *DeploymentResource) checkSameNamespace(namespace, serviceName string) bool {
	svc, err := model.GetServiceInfo(serviceName)
	if err != nil {
		log.Errorf("[deployment] query service by id: %s failed: %s", serviceName, err)
		return false
	}
	if namespace != svc.Namespace {
		log.Errorf("[deployment] service: %s namespace: %s != %s", serviceName, svc.Namespace, namespace)
		return false
	}
	return true
}

func (r *DeploymentResource) checkPhaseFinish(pipelineID int64, kind, phase string) bool {
	ph, err := model.GetPhaseInfo(pipelineID, kind, phase)
	if errors.Is(err, model.NotFound) {
		return true
	} else if err != nil {
		log.Errorf("[deployment] query phase info error: %s", err)
		return true
	}

	// 判断该阶段是否完成
	if ph.Status == model.PHSuccess {
		return true
	}
	return false
}
