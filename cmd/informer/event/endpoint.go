package event

import (
	"regexp"
	"sort"
	"strconv"
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

	ips, ready := r.parseIP(subsets)
	if !ready {
		return nil
	}

	serviceID, serviceName, phase, err := r.parseInfo(name)
	if err != nil {
		return err
	}
	log.Infof("[endpoint] service: %s phase: %s have total ips: %#v", serviceName, phase, ips)

	svc, err := model.GetServiceByID(serviceID)
	if err != nil {
		log.Errorf("[endpoint] query service by id error: %+v", err)
		return err
	}
	ns, err := model.GetNamespaceByID(svc.NamespaceID)
	if err != nil {
		log.Errorf("[endpoint] query namespace by id error: %+v", err)
		return err
	}
	if namespace != ns.Name {
		log.Errorf("[endpoint] service: %s namespace: %s != %s", serviceName, ns.Name, namespace)
		return nil
	}

	if err := r.updateTraffic(serviceName, phase, ips); err != nil {
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

func (r *EndpointResource) parseIP(subsets []corev1.EndpointSubset) ([]string, bool) {
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

func (r *EndpointResource) parseInfo(name string) (int64, string, string, error) {
	// 获取服务ID
	re := regexp.MustCompile(`-\d+-`)
	result := re.FindStringSubmatch(name)
	match := strings.Trim(result[0], "-")
	serviceID, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		log.Errorf("[endpoint] parse: %s convert to int64 error: %s", name, err)
		return 0, "", "", err
	}

	// 获取服务名
	matchList := re.Split(name, -1)
	serviceName := matchList[0]
	phase := matchList[1]
	return serviceID, serviceName, phase, nil
}

func (r *EndpointResource) updateTraffic(service, phase string, ips []string) error {
	log.Infof("[endpoint] service: %s phase: %s update traffic: %#v success", service, phase, ips)
	return nil
}
