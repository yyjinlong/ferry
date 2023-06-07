package event

import (
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/pkg/model"
)

type Endpoint interface {
	HandleEndpoint(obj interface{}, mode, cluster string) error
}

type EndpointResource struct {
	clientset *kubernetes.Clientset
}

func NewEndpointResource(clientset *kubernetes.Clientset) *EndpointResource {
	return &EndpointResource{
		clientset: clientset,
	}
}

func (r *EndpointResource) HandleEndpoint(obj interface{}, mode, cluster string) error {
	var (
		data      = obj.(*corev1.Endpoints)
		name      = data.ObjectMeta.Name
		namespace = data.ObjectMeta.Namespace
		subsets   = data.Subsets
	)

	if !r.filter(name) {
		return nil
	}

	ips, ready := r.parseAddr(subsets)
	if !ready {
		return nil
	}

	serviceName, phase, group := r.parseInfo(name)
	log.Infof("[endpoint] service: %s phase: %s group: %s have total ips: %#v", serviceName, phase, group, ips)

	svc, err := model.GetServiceInfo(serviceName)
	if err != nil {
		log.Errorf("[endpoint] query service by id error: %+v", err)
		return err
	}

	var (
		curNamespace = svc.Namespace
		deployGroup  = svc.DeployGroup
		onlineGroup  = svc.OnlineGroup
	)

	if namespace != curNamespace {
		log.Errorf("[endpoint] service: %s namespace: %s != %s", serviceName, curNamespace, namespace)
		return nil
	}

	if err := r.traffic(namespace, serviceName, phase, group, deployGroup, onlineGroup, ips); err != nil {
		log.Errorf("[endpoint] update service: %s traffic failed: %+v", serviceName, err)
		return err
	}
	return nil
}

func (r *EndpointResource) filter(name string) bool {
	re := regexp.MustCompile(`[\w+-]+-\d+-\w+`)
	if re == nil {
		return false
	}

	result := re.FindAllStringSubmatch(name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (r *EndpointResource) parseInfo(name string) (string, string, string) {
	re := regexp.MustCompile(`-\d+-`)
	matchList := re.Split(name, -1)

	// 获取服务名
	serviceName := matchList[0]

	// 获取阶段、部署组
	other := matchList[1]
	otherList := strings.Split(other, "-")
	phase := otherList[0]
	group := otherList[1]
	return serviceName, phase, group
}

func (r *EndpointResource) parseAddr(subsets []corev1.EndpointSubset) ([]string, bool) {
	ready := false
	ips := make([]string, 0)
	for _, item := range subsets {
		// NotReadyAddresses list is empty, indicate pod is ready
		if len(item.NotReadyAddresses) == 0 {
			ready = true
		}

		for _, addr := range item.Addresses {
			ips = append(ips, addr.IP)
		}
	}
	sort.Strings(ips)
	return ips, ready
}

func (r *EndpointResource) traffic(namespace, service, phase, group, deployGroup, onlineGroup string, ips []string) error {
	// 两组(blue、green) 只有一组接流量
	pipeline, err := model.GetServicePipeline(service)
	if err != nil {
		log.Errorf("get service: %s pipeline info failed: %+v", service, err)
		return err
	}
	pipelineID := pipeline.ID

	if pipeline.Status == model.PLProcess {
		// NOTE: 发布中
		if group == deployGroup {
			// 发布中且该组是发布组, 则更新流量
			r.update(service, phase, group, ips)
			return nil

		} else {
			// 发布中且该组不是发布组
			// 判断该阶段是否已发布完成; 如果没有, 则表示是需要更新该组的流量
			// case: 正在发布green组sandbox, 且green组online没有发布. 这时blue组online pod变化了, 则需要更新
			if !model.CheckPhaseIsDeploy(pipelineID, model.PHASE_DEPLOY, phase) {
				r.update(service, phase, group, ips)
				return nil
			}
		}

	} else if pipeline.Status == model.PLRollbacking {
		// NOTE: 回滚中
		if !model.CheckDeployFinish(pipelineID) {
			// 发布中回滚(当前组为在线组)
			if group == onlineGroup {
				r.update(service, phase, group, ips)
				return nil
			}

		} else {
			// 发布完成回滚(当前组为部署组)
			if group == deployGroup {
				r.update(service, phase, group, ips)
				return nil
			}
		}

	} else {
		// NOTE: 发布成功、失败(只更新对应阶段、对应组的流量)
		r.update(service, phase, group, ips)
	}
	return nil
}

func (r *EndpointResource) update(service, phase, group string, ips []string) {
	log.Infof("[endpoint] service: %s phase: %s group: %s update traffic: %#v success", service, phase, group, ips)
}
