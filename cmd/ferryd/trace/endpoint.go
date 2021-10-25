// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package trace

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ferry/pkg/g"
	"ferry/pkg/log"
	"ferry/server/db"
	"ferry/server/objects"
)

var (
	endResourceVersionMap = make(map[string]string)
)

func handleEndpoint(obj interface{}, mode string) {
	data := obj.(*corev1.Endpoints)
	ept := &endpoint{
		mode:            mode,
		service:         data.Name,
		subsets:         data.Subsets,
		resourceVersion: data.ObjectMeta.ResourceVersion,
	}
	ept.worker()
}

type endpoint struct {
	mode            string
	service         string
	serviceID       int64
	serviceName     string
	subsets         []corev1.EndpointSubset
	resourceVersion string
}

func (e *endpoint) worker() {
	if !e.isValidService() {
		return
	}
	if !e.parseServiceID() {
		return
	}
	if !e.parseServiceName() {
		return
	}

	key := e.serviceName
	version, ok := endResourceVersionMap[key]
	if ok && version == e.resourceVersion {
		log.Infof("[%s] service: %s resource version same so stop", e.mode, e.service)
		return
	}
	endResourceVersionMap[key] = e.resourceVersion

	pipeline, err := objects.GetServicePipeline(e.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] service: %s get pipeline id error: %s", e.mode, e.service, err)
		return
	}
	if g.Ini(pipeline.Pipeline.Status, []int{db.PLSuccess, db.PLRollbackSuccess}) {
		delete(endResourceVersionMap, key) // NOTE: 上线完成删除对应的key
		log.Infof("[%s] service: %s deploy finish so stop", e.mode, e.service)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipeline.Pipeline.ID,
		"service":    e.service,
		"new":        e.getIPList(),
	}
	log.Infof("[%s] service: %s endpoints: %+v", e.mode, e.service, result)
}

func (e *endpoint) isValidService() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(e.service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (e *endpoint) parseServiceID() bool {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return false
	}

	result := reg.FindAllStringSubmatch(e.service, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("[%s] service: %s convert to int64 error: %s", e.mode, e.service, err)
		return false
	}
	e.serviceID = serviceID
	return true
}

func (e *endpoint) parseServiceName() bool {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return false
	}
	matchList := reg.Split(e.service, -1)
	e.serviceName = matchList[0]
	return true
}

func (e *endpoint) getIPList() []string {
	ipList := make([]string, 0)
	for _, item := range e.subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
