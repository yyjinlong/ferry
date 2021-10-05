// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package handle

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ferry/ops/objects"
)

func NewEndpointEvent(obj interface{}, mode string) *EndpointEvent {
	data := obj.(*corev1.Endpoints)
	return &EndpointEvent{
		mode:    mode,
		service: data.Name,
		subsets: data.Subsets,
	}
}

type EndpointEvent struct {
	mode    string
	service string
	subsets []corev1.EndpointSubset
}

func (ee *EndpointEvent) IsValidService() bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(ee.service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (ee *EndpointEvent) GetIPList() []map[string]string {
	ipList := make([]map[string]string, 0)
	for _, item := range ee.subsets {
		for _, addrInfo := range item.Addresses {
			ipInfo := make(map[string]string)
			ipInfo["podIp"] = addrInfo.IP
			ipInfo["podName"] = addrInfo.TargetRef.Name
			ipList = append(ipList, ipInfo)
		}
	}
	return ipList
}

func (ee *EndpointEvent) Parse() (int64, error) {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return 0, fmt.Errorf("regexp compile error")
	}

	result := reg.FindAllStringSubmatch(ee.service, -1)
	if len(result) == 0 {
		return 0, fmt.Errorf("regexp is not match")
	}

	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	pipeline, err := objects.GetServicePipeline(serviceID)
	if err != nil {
		return 0, err
	}
	return pipeline.Pipeline.ID, nil
}

func (ee *EndpointEvent) GetService() string {
	return ee.service
}
