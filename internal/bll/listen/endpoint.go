// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

func CheckEndpointIsFinish(newObj interface{}, oldObj interface{}, mode string) {
	oldSubsets := make([]corev1.EndpointSubset, 0)
	if oldObj != nil {
		oldData := oldObj.(*corev1.Endpoints)
		oldSubsets = oldData.Subsets
	}

	newData := newObj.(*corev1.Endpoints)
	ef := &endpointFinish{
		mode:            mode,
		service:         newData.Name,
		newSubsets:      newData.Subsets,
		oldSubsets:      oldSubsets,
		resourceVersion: newData.ObjectMeta.ResourceVersion,
	}
	ef.worker()
}

type endpointFinish struct {
	mode            string
	service         string
	serviceID       int64
	serviceName     string
	newSubsets      []corev1.EndpointSubset
	oldSubsets      []corev1.EndpointSubset
	resourceVersion string
}

func (ef *endpointFinish) worker() {
	log.InitFields(log.Fields{
		"logid": g.UniqueID(), "mode": ef.mode, "type": "endpoint", "version": ef.resourceVersion})
	if !ef.isValidService() {
		return
	}
	if !ef.parseServiceName() {
		return
	}
	if !ef.parseServiceID() {
		return
	}

	key := ef.serviceName
	version, ok := endResourceVersionMap[key]
	if ok && version == ef.resourceVersion {
		log.Infof("service: %s resource version same so stop", ef.service)
		return
	}
	endResourceVersionMap[key] = ef.resourceVersion

	pipeline, err := objects.GetServicePipeline(ef.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("service: %s get pipeline id error: %s", ef.service, err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		delete(endResourceVersionMap, key) // NOTE: 上线完成删除对应的key
		log.Infof("service: %s deploy finish so stop", ef.service)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipeline.Pipeline.ID,
		"service":    ef.service,
		"online":     ef.getIPList(ef.newSubsets),
		"offline":    ef.getIPList(ef.oldSubsets),
	}
	log.Infof("service: %s endpoints: %+v", ef.service, result)
}

func (ef *endpointFinish) isValidService() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(ef.service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (ef *endpointFinish) parseServiceName() bool {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return false
	}
	matchList := reg.Split(ef.service, -1)
	ef.serviceName = matchList[0]
	return true
}

func (ef *endpointFinish) parseServiceID() bool {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return false
	}

	result := reg.FindAllStringSubmatch(ef.service, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("service: %s convert to int64 error: %s", ef.service, err)
		return false
	}
	ef.serviceID = serviceID
	return true
}

func (ef *endpointFinish) getIPList(subsets []corev1.EndpointSubset) []string {
	ipList := make([]string, 0)
	for _, item := range subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
