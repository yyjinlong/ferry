package event

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/golib/log"
	"nautilus/pkg/k8s/client"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func HandleEndpointCapturer(obj interface{}, mode string, clientset *kubernetes.Clientset) {
	var (
		data = obj.(*corev1.Endpoints)
		name = data.ObjectMeta.Name
	)

	log.InitFields(log.Fields{
		"mode":     mode,
		"endpoint": name,
		"version":  data.ObjectMeta.ResourceVersion,
	})

	handleEvent(&endpointCapturer{
		mode:      mode,
		name:      name,
		subsets:   data.Subsets,
		clientset: clientset,
	})
}

type endpointCapturer struct {
	mode        string
	name        string
	subsets     []corev1.EndpointSubset
	clientset   *kubernetes.Clientset
	serviceID   int64
	serviceName string
	phase       string
}

func (e *endpointCapturer) valid() bool {
	// NOTE: 检查是否是业务的service
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

func (e *endpointCapturer) ready() bool {
	return true
}

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
	currentService := matchList[0]

	afterList := strings.Split(matchList[1], "-")
	currentPhase := afterList[0]
	if currentService == service && currentPhase == phase {
		return true
	}
	return false
}
