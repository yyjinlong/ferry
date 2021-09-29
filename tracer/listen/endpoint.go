// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	corev1 "k8s.io/api/core/v1"

	"ferry/ops/g"
	"ferry/ops/log"
)

func endpointAdd(obj interface{}) {
	data := obj.(*corev1.Endpoints)
	service := data.Name
	if ignore(service) {
		return
	}
	ipList := addresses(data.Subsets)
	log.Infof("service: %s added, endpoint: %+v", service, ipList)
}

func endpointUpdate(oldObj, newObj interface{}) {
	oldData := oldObj.(*corev1.Endpoints)
	newData := newObj.(*corev1.Endpoints)

	service := oldData.Name
	if ignore(service) {
		return
	}

	oldIpList := addresses(oldData.Subsets)
	log.Infof("service: %s update, old endpoint: %+v", service, oldIpList)

	newIpList := addresses(newData.Subsets)
	log.Infof("service: %s update, new endpoint: %+v", service, newIpList)
	// TODO: deployment完成后，将endpoint记录, 写到事件表里, 并对外提供接口获取最新节点的ip列表.
}

func endpointDelete(obj interface{}) {
	data := obj.(*corev1.Endpoints)
	service := data.Name
	if ignore(service) {
		return
	}
	ipList := addresses(data.Subsets)
	log.Infof("service: %s delete, old endpoint: %+v", service, ipList)
}

func ignore(name string) bool {
	list := []string{"kubernetes", "kube-scheduler", "kube-controller-manager", "cattle-cluster-agent"}
	if g.In(name, list) {
		return true
	}
	return false
}

func addresses(subsets []corev1.EndpointSubset) []map[string]string {
	ipList := make([]map[string]string, 0)
	for _, item := range subsets {
		ipInfo := make(map[string]string)
		for _, addrInfo := range item.Addresses {
			ipInfo["podIp"] = addrInfo.IP
			ipInfo["podName"] = addrInfo.TargetRef.Name
			ipList = append(ipList, ipInfo)
		}
	}
	return ipList
}
