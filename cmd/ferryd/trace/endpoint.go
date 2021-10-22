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

func handleEndpoint(oldObj, newObj interface{}, mode string) {
	var (
		data    = newObj.(*corev1.Endpoints)
		service = data.Name
		subsets = data.Subsets
	)

	if !isValidService(service) {
		return
	}

	pipelineID, err := getPipelineID(service)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("[%s] get pipeline id error: %s", mode, err)
		return
	}

	switch mode {
	case Create:
		wrap(pipelineID, service, getIPList(subsets), nil)
	case Update:
		oldData := oldObj.(*corev1.Endpoints)
		wrap(pipelineID, service, getIPList(subsets), getIPList(oldData.Subsets))
	case Delete:
		wrap(pipelineID, service, nil, getIPList(subsets))
	}
}

func wrap(pipelineID int64, service string, addList, delList []string) map[string]interface{} {
	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    service,
		"add":        addList,
		"del":        delList,
	}
	log.Infof("endpoint result: %+v", result)
	return result
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
