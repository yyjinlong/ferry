package event

import (
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/golib/log"
)

func HandleEndpointCapturer(obj interface{}, mode string, clientset *kubernetes.Clientset) {
	var (
		data = obj.(*corev1.Endpoints)
		name = data.ObjectMeta.Name
	)

	log.InitFields(log.Fields{
		"mode":    mode,
		"name":    name,
		"version": data.ObjectMeta.ResourceVersion,
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
	return true
}

func (e *endpointCapturer) getIPList() []string {
	ips := make([]string, 0)
	for _, item := range e.subsets {
		for _, addrInfo := range item.Addresses {
			ips = append(ips, addrInfo.IP)
		}
	}
	return ips
}

func GetServicePods(clientset *kubernetes.Clientset, namespace, service, phase string) []string {
	var ips []string
	pod, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return ips
	}

	for _, obj := range pod.Items {
		name := obj.ObjectMeta.Name
		podIP := obj.Status.PodIP
		status := obj.Status.Phase

		if isCurrentService(name, service, phase) && status == corev1.PodRunning {
			ips = append(ips, podIP)
		}
	}
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
