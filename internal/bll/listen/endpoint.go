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

func CheckEndpointIsFinish(obj interface{}, mode string) {
	data := obj.(*corev1.Endpoints)
	ef := &endpointFinish{
		mode:            mode,
		service:         data.Name,
		subsets:         data.Subsets,
		resourceVersion: data.ObjectMeta.ResourceVersion,
	}
	ef.worker()
}

type endpointFinish struct {
	mode            string
	service         string
	serviceID       int64
	serviceName     string
	subsets         []corev1.EndpointSubset
	resourceVersion string
}

func (ef *endpointFinish) worker() {
	if !ef.isValidService() {
		return
	}
	if !ef.parseServiceID() {
		return
	}
	if !ef.parseServiceName() {
		return
	}

	key := ef.serviceName
	version, ok := endResourceVersionMap[key]
	if ok && version == ef.resourceVersion {
		log.Infof("[%s] service: %s resource version same so stop", ef.mode, ef.service)
		return
	}
	endResourceVersionMap[key] = ef.resourceVersion

	pipeline, err := objects.GetServicePipeline(ef.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] service: %s get pipeline id error: %s", ef.mode, ef.service, err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		delete(endResourceVersionMap, key) // NOTE: 上线完成删除对应的key
		log.Infof("[%s] service: %s deploy finish so stop", ef.mode, ef.service)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipeline.Pipeline.ID,
		"service":    ef.service,
		"new":        ef.getIPList(),
	}
	log.Infof("[%s] service: %s endpoints: %+v", ef.mode, ef.service, result)
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
		log.Errorf("[%s] service: %s convert to int64 error: %s", ef.mode, ef.service, err)
		return false
	}
	ef.serviceID = serviceID
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

func (ef *endpointFinish) getIPList() []string {
	ipList := make([]string, 0)
	for _, item := range ef.subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
