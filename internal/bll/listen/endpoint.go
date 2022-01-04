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

func CheckEndpointIsFinish(newObj interface{}, mode string) {
	var (
		newData         = newObj.(*corev1.Endpoints)
		service         = newData.Name
		resourceVersion = newData.ObjectMeta.ResourceVersion
	)

	log.InitFields(log.Fields{
		"logid":   g.UniqueID(),
		"mode":    mode,
		"service": service,
		"version": resourceVersion,
	})

	eh := &endhandler{
		mode:            mode,
		service:         service,
		newSubsets:      newData.Subsets,
		resourceVersion: resourceVersion,
	}
	if !eh.valid() {
		return
	}
	if !eh.parse() {
		return
	}
	if !eh.operate() {
		return
	}
}

type endhandler struct {
	mode            string
	service         string
	phase           string
	serviceID       int64
	serviceName     string
	resourceVersion string
	newSubsets      []corev1.EndpointSubset
}

func (h *endhandler) valid() bool {
	// 检查是否是业务的service
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}

	result := reg.FindAllStringSubmatch(h.service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (h *endhandler) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}

	// 获取服务ID
	result := reg.FindAllStringSubmatch(h.service, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service: %s convert to int64 error: %s", h.service, err)
		return false
	}
	h.serviceID = serviceID

	// 获取服务名
	matchList := reg.Split(h.service, -1)
	h.serviceName = matchList[0]
	h.phase = matchList[1]
	return true
}

func (h *endhandler) operate() bool {
	pipeline, err := objects.GetServicePipeline(h.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service id: %d error: %s", h.serviceID, err)
		return false
	}

	// 判断该上线流程是否完成
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Infof("check service: %s deploy finish so stop", h.service)
		return false
	}

	pipelineID := pipeline.Pipeline.ID
	phaseList, err := objects.FindPhases(pipelineID)
	if err != nil {
		log.Errorf("query pipeline phases error: %s", err)
		return false
	}
	lastPhase := phaseList[0]

	// 判断该阶段是否完成
	if g.Ini(lastPhase.Status, []int{model.PHSuccess, model.PHFailed}) {
		return true
	}

	// 拼成最后结果
	result := map[string]interface{}{
		"pipelineID": pipelineID,
		"service":    h.serviceName,
		"phase":      h.phase,
		"online":     h.getIPList(h.newSubsets),
	}
	log.Infof("service: %s endpoints: %+v", h.service, result)
	return true
}

func (h *endhandler) getIPList(subsets []corev1.EndpointSubset) []string {
	ipList := make([]string, 0)
	for _, item := range subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
