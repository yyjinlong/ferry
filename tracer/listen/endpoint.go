// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ferry/ops/log"
	"ferry/ops/objects"
)

func EndpointAdd(obj interface{}) {
	epEvent := NewEndpointEvent(obj, "add")
	if !epEvent.IsValidService() {
		return
	}
	pipelineID, err := epEvent.Parse()
	if err != nil {
		log.Errorf("parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    epEvent.GetService(),
		"add":        epEvent.GetIPList(),
		"del":        make([]map[string]string, 0),
	}
	log.Infof("endpoint add result: %+v", result)
}

func EndpointUpdate(oldObj, newObj interface{}) {
	oldEvent := NewEndpointEvent(oldObj, "update")
	if !oldEvent.IsValidService() {
		return
	}
	newEvent := NewEndpointEvent(newObj, "update")
	if !newEvent.IsValidService() {
		return
	}
	pipelineID, err := newEvent.Parse()
	if err != nil {
		log.Errorf("parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    newEvent.GetService(),
		"add":        newEvent.GetIPList(),
		"del":        oldEvent.GetIPList(),
	}
	log.Infof("endpoint update result: %+v", result)
}

func EndpointDelete(obj interface{}) {
	epEvent := NewEndpointEvent(obj, "delete")
	if !epEvent.IsValidService() {
		return
	}
	pipelineID, err := epEvent.Parse()
	if err != nil {
		log.Errorf("parser pipeline id error: %s", err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    epEvent.GetService(),
		"add":        make([]map[string]string, 0),
		"del":        epEvent.GetIPList(),
	}
	log.Infof("endpoint delete result: %+v", result)
}

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
