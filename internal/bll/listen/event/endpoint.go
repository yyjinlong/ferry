// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

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

func HandleEndpointCapturer(obj interface{}, mode string) {
	var (
		data            = obj.(*corev1.Endpoints)
		service         = data.Name
		resourceVersion = data.ObjectMeta.ResourceVersion
	)

	log.InitFields(log.Fields{
		"logid":   g.UniqueID(),
		"mode":    mode,
		"service": service,
		"version": resourceVersion,
	})

	handleEvent(&endpointCapturer{
		mode:            mode,
		service:         service,
		newSubsets:      data.Subsets,
		resourceVersion: resourceVersion,
	})
}

type endpointCapturer struct {
	mode            string
	service         string
	phase           string
	serviceID       int64
	serviceName     string
	resourceVersion string
	newSubsets      []corev1.EndpointSubset
}

func (c *endpointCapturer) valid() bool {
	// 检查是否是业务的service
	reg := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if reg == nil {
		return false
	}

	result := reg.FindAllStringSubmatch(c.service, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (c *endpointCapturer) ready() bool {
	return true
}

func (c *endpointCapturer) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)
	if reg == nil {
		return false
	}

	// 获取服务ID
	result := reg.FindAllStringSubmatch(c.service, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service: %s convert to int64 error: %s", c.service, err)
		return false
	}
	c.serviceID = serviceID

	// 获取服务名
	matchList := reg.Split(c.service, -1)
	c.serviceName = matchList[0]
	c.phase = matchList[1]
	return true
}

func (c *endpointCapturer) operate() bool {
	pipeline, err := objects.GetServicePipeline(c.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service id: %d error: %s", c.serviceID, err)
		return false
	}

	// 判断该上线流程是否完成
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Infof("check service: %s deploy finish so stop", c.service)
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
		"service":    c.serviceName,
		"phase":      c.phase,
		"online":     c.getIPList(c.newSubsets),
	}
	log.Infof("service: %s endpoints: %+v", c.service, result)
	return true
}

func (c *endpointCapturer) getIPList(subsets []corev1.EndpointSubset) []string {
	ipList := make([]string, 0)
	for _, item := range subsets {
		for _, addrInfo := range item.Addresses {
			ipList = append(ipList, addrInfo.IP)
		}
	}
	return ipList
}
