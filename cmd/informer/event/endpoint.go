package event

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/pkg/k8s/client"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func (e *endpointCapturer) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}

	// 获取服务ID
	result := reg.FindAllStringSubmatch(e.name, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse endpoint: %s convert to int64 error: %s", e.name, err)
		return false
	}
	e.serviceID = serviceID

	// 获取服务名
	matchList := reg.Split(e.name, -1)
	e.serviceName = matchList[0]
	e.phase = matchList[1]
	return true
}

func (e *endpointCapturer) operate() bool {
	ips := e.getIPList()
	log.Infof("service: %s phase: %s have total ips: %#v", e.serviceName, e.phase, ips)

	serviceObj, err := model.GetServiceByID(e.serviceID)
	if err != nil {
		log.Errorf("query service by id error: %+v", err)
		return false
	}
	namespaceObj, err := model.GetNamespace(serviceObj.NamespaceID)
	if err != nil {
		log.Errorf("query namespace by id error: %+v", err)
		return false
	}
	namespace := namespaceObj.Name

	switch e.phase {
	case model.PHASE_SANDBOX:
		if len(ips) == 1 && e.checkPodReady(namespace, ips) {
			log.Infof("-------service: %s sandbox ready ips: %#v", e.serviceName, ips)
			return true
		}

	case model.PHASE_ONLINE:
		if len(ips) == serviceObj.Replicas && e.checkPodReady(namespace, ips) {
			log.Infof("-------service: %s online ready ips: %#v", e.serviceName, ips)
			return true
		}
	}
	return false
}

func (e *endpointCapturer) getIPList() []string {
	ips := make([]string, 0)
	for _, item := range e.subsets {
		for _, addrInfo := range item.Addresses {
			ips = append(ips, addrInfo.IP)
		}
	}
	sort.Strings(ips)
	return ips
}

func (e *endpointCapturer) checkPodReady(namespace string, ips []string) bool {
	podIPs := GetServicePods(e.clientset, namespace, e.serviceName, e.phase)
	for _, ip := range ips {
		if !util.In(ip, podIPs) {
			log.Warnf("service: %s phase: %s ip: %s not in pod list", e.serviceName, e.phase, ip)
			return false
		}
	}
	log.Infof("service: %s phase: %s endpoint listen all pod are ready", e.serviceName, e.phase)
	return true
}

func GetServicePods(clientset *kubernetes.Clientset, namespace, service, phase string) []string {
	var ips []string
	pods, err := client.GetPods(clientset, namespace)
	if err != nil {
		return ips
	}

	for _, obj := range pods.Items {
		name := obj.ObjectMeta.Name
		status := obj.Status.Phase
		if isCurrentService(name, service, phase) && status == corev1.PodRunning {
			ips = append(ips, obj.Status.PodIP)
		}
	}
	sort.Strings(ips)
	return ips
}

func isCurrentService(podName, service, phase string) bool {
	reg := regexp.MustCompile(`-\d+-`)
	matchList := reg.Split(podName, -1)
	if len(matchList) < 2 {
		return false
	}

	currentService := matchList[0]
	afterList := strings.Split(matchList[1], "-")
	currentPhase := afterList[0]
	if currentService == service && currentPhase == phase {
		return true
	}
	return false
}

const (
	ENDPOINT_KEY = "redispaas:distributed:endpoint"
	ENDPOINT_VAL = "endpoint"
)

type Endpoint interface {
	HandleEndpoint(obj interface{}, mode, clusterName string) error
}

type EndpointResource struct {
	clientset *kubernetes.Clientset
}

func NewEndpointResource(clientset *kubernetes.Clientset) *EndpointResource {
	return &EndpointResource{
		clientset: clientset,
	}
}

func (e *EndpointResource) HandleEndpoint(obj interface{}, mode, clusterName string) error {
	var (
		data    = obj.(*corev1.Endpoints)
		name    = data.ObjectMeta.Name
		subsets = data.Subsets
	)

	if !e.filter(name) {
		return nil
	}

	ips, ready := e.parseIP(subsets)
	if !ready {
		return nil
	}

	// 获取分布式锁, 超时时间为20秒.
	rs := session.NewRedisService()
	rs.Connect(
		config.Config().Redis.Addr,
		config.Config().Redis.Password,
		config.Config().Redis.DB,
		config.Config().Redis.Pool,
	)
	if rs.AcquireLock(ENDPOINT_KEY, ENDPOINT_VAL, 20) {
		defer func() {
			rs.ReleaseLock(ENDPOINT_KEY)
		}()

		component, serviceName, err := e.parseName(name)
		if err != nil {
			log.Errorf("redis cluster: %s parse endpoint name error: %+v", name, err)
			return err
		}
		log.Infof("redis cluster: %s current endpoint get change ips: %v", name, ips)

		// 获取到锁, 进行流量更新
		cluster := service.NewCluster()
		treenode := cluster.GenerateTreeNode(serviceName, component)
		if err := cluster.UpdateTraffic(treenode, clusterName, component, serviceName, ips); err != nil {
			log.Errorf("redis cluster: %s update traffic failed: %+v", name, err)
			return err
		}
		log.Infof("redis cluster: %s update traffic ips: %v success", name, ips)

		if component == "predixy" {
			if err := model.UpdateClusterStatus(serviceName, model.CLUSTER_SUCCESS); err != nil {
				log.Errorf("redis cluster: %s update cluster status failed: %+v", name, err)
				return err
			}
			log.Infof("redis cluster: %s update cluster status success", name)
		}
	}
	return nil
}

func (e *EndpointResource) filter(name string) bool {
	// 检查是否是业务的service
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(e.name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (e *EndpointResource) parseIP(subsets []corev1.EndpointSubset) ([]string, bool) {
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

// parse endpoint name return component and service
func (e *EndpointResource) parseName(name string) (string, string, error) {
	segments := strings.SplitN(name, "-", 3)
	if len(segments) < 3 {
		return "", "", fmt.Errorf("error endpoint name: %s", name)
	}

	prefix := segments[0]
	component := PREFIX_MAP[prefix]
	return component, segments[2], nil
}
