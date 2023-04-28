// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"errors"
	"regexp"
	"strconv"
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

	// 检查deployment是否ready
	if !(data.ObjectMeta.Generation == data.Status.ObservedGeneration &&
		replicas == updatedReplicas &&
		replicas == availableReplicas &&
		replicas == readyReplicas) {
		return nil
	}

	serviceID, serviceName, phase, group, err := r.parseInfo(name)
	if err != nil {
		return err
	}
	log.Infof("[deployment] check: %s ready with group: %s replicas: %d", name, group, replicas)

	pipeline, err := model.GetServicePipeline(serviceID)
	if !errors.Is(err, model.NotFound) && err != nil {
		log.Errorf("[deployment] query pipeline by service id: %d error: %s", serviceID, err)
		return err
	}
	pipelineID := pipeline.ID

	if cm.Ini(pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Info("[deployment] check deploy is finished")
		return err
	}

	svc, err := model.GetServiceByID(serviceID)
	if err != nil {
		log.Errorf("[deployment] query service by id: %d failed: %s", serviceID, err)
		return err
	}
	namespaceID := svc.NamespaceID

	ns, err := model.GetNamespaceByID(namespaceID)
	if err != nil {
		log.Errorf("[deployment] query namespace by id: %d failed: %s", namespaceID, err)
		return err
	}

	if namespace != ns.Name {
		log.Errorf("[deployment] service: %s namespace: %s != %s", serviceName, ns.Name, namespace)
		return nil
	}

	kind := model.PHASE_DEPLOY
	if pipeline.Status == model.PLRollbacking {
		kind = model.PHASE_ROLLBACK
	}
	log.Infof("[deployment] get pipeline: %d kind: %s phase: %s", pipelineID, kind, phase)

	ph, err := model.GetPhaseInfo(pipelineID, kind, phase)
	if errors.Is(err, model.NotFound) {
		return nil
	} else if err != nil {
		log.Errorf("[deployment] query phase info error: %s", err)
		return err
	}

	// 判断该阶段是否完成
	if ph.Status == model.PHSuccess {
		return nil
	}

	// 获取endpoint信息

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

func (r *DeploymentResource) parseInfo(name string) (int64, string, string, string, error) {
	re := regexp.MustCompile(`-\d+-`)

	// 获取服务名、阶段、蓝绿组
	matchList := re.Split(name, -1)
	serviceName := matchList[0]

	afterList := strings.Split(matchList[1], "-")
	phase := afterList[0]
	group := afterList[1]

	// 获取服务ID
	result := re.FindStringSubmatch(name)
	match := strings.Trim(result[0], "-")
	serviceID, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		log.Errorf("[deployment] parse: %s convert to int64 error: %s", name, err)
		return 0, "", "", "", err
	}
	return serviceID, serviceName, phase, group, nil
}
