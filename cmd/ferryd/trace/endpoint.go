// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package trace

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ferry/pkg/log"
	"ferry/server/objects"
)

func handleEndpoint(obj interface{}, mode string) {
	var (
		data    = obj.(*corev1.Endpoints)
		service = data.Name
		subsets = data.Subsets
	)

	if !isValidService(service) {
		return
	}

	pipelineID, err := getPipelineID(service)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] service: %s get pipeline id error: %s", mode, service, err)
		return
	}

	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    service,
		"new":        getIPList(subsets),
	}
	log.Infof("[%s] service: %s endpoints: %+v", mode, service, result)
}

func isValidService(service string) bool {
	reg := regexp.MustCompile(`[\w+-]+-\d+`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func getPipelineID(service string) (int64, error) {
	reg := regexp.MustCompile(`-\d+`)
	if reg == nil {
		return 0, fmt.Errorf("regexp compile error")
	}

	result := reg.FindAllStringSubmatch(service, -1)
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

func getIPList(subsets []corev1.EndpointSubset) []string {
	ipList := make([]string, 0)
	for _, item := range subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
