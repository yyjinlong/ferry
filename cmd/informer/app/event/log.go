// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"nautilus/golib/log"
	"nautilus/pkg/model"
	"nautilus/pkg/util"
)

func HandleLogCapturer(obj interface{}, mode string) {
	data := obj.(*corev1.Event)
	var (
		message = data.Message
		name    = data.Name
		fields  = data.ObjectMeta.ManagedFields
	)

	log.InitFields(log.Fields{"mode": mode, "event": name})

	handleEvent(&logCapturer{
		name:    name,
		message: message,
		fields:  fields,
	})
}

type logCapturer struct {
	name      string
	message   string
	fields    []metav1.ManagedFieldsEntry
	service   string
	serviceID int64
	phase     string
}

func (c *logCapturer) valid() bool {
	re := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if re == nil {
		return false
	}
	result := re.FindAllStringSubmatch(c.name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (c *logCapturer) ready() bool {
	return true
}

func (c *logCapturer) parse() bool {
	re := regexp.MustCompile(`-\d+-`)
	if re == nil {
		return false
	}

	matchList := re.Split(c.name, -1)
	c.service = matchList[0]

	afterList := strings.Split(matchList[1], "-")
	c.phase = afterList[0]

	result := re.FindAllStringSubmatch(c.name, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service id convert to int64 error: %s", err)
		return false
	}
	c.serviceID = serviceID
	return true
}

func (c *logCapturer) operate() bool {
	pipeline, err := model.GetServicePipeline(c.serviceID)
	if !errors.Is(err, model.NotFound) && err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return false
	}
	pipelineID := pipeline.ID

	// 判断该上线流程是否完成
	if util.Ini(pipeline.Status, []int{model.PLSuccess, model.PLRollbacking, model.PLRollbackSuccess}) {
		log.Info("check deploy is finished, stop update event log")
		return false
	}

	kind := model.PHASE_DEPLOY
	if util.Ini(pipeline.Status, []int{model.PLRollbacking, model.PLRollbackSuccess, model.PLRollbackFailed}) {
		kind = model.PHASE_ROLLBACK
	}

	log.Infof("get pipeline: %d kind: %s phase: %s", pipelineID, kind, c.phase)

	info := c.fields[0]
	operTime := info.Time.Format("15:04:05")

	re := regexp.MustCompile(`[\w+-]+-\d+-\w+-\w+`)
	result := re.FindAllStringSubmatch(c.name, -1)
	if len(result) == 0 {
		return false
	}
	operName := result[0][0]

	msg := fmt.Sprintf("[%s] %v\n%s", operTime, operName, c.message)
	err = model.RealtimeLog(pipelineID, kind, c.phase, msg)
	if !errors.Is(err, model.NotFound) && err != nil {
		log.Errorf("write event log to db error: %s", err)
		return false
	}
	return true
}
