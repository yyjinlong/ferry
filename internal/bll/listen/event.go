// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ferry/internal/model"
	"ferry/internal/objects"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

func FetchPublishEvent(obj interface{}, mode string) {
	var (
		data       = obj.(*corev1.Event)
		objectMeta = data.ObjectMeta
		message    = data.Message
		name       = data.Name
		fields     = objectMeta.ManagedFields
	)

	log.InitFields(log.Fields{
		"logid": g.UniqueID(),
		"mode":  mode,
		"event": name,
	})

	pe := &publishEvent{
		name:    name,
		message: message,
		fields:  fields,
	}
	if !pe.valid() {
		return
	}
	if !pe.parse() {
		return
	}
	if !pe.operate() {
		return
	}
}

type publishEvent struct {
	name      string
	message   string
	fields    []metav1.ManagedFieldsEntry
	service   string
	serviceID int64
	phase     string
}

func (pe *publishEvent) valid() bool {
	re := regexp.MustCompile(`[\w+-]+-\d+-[\w+-]+`)
	if re == nil {
		return false
	}
	result := re.FindAllStringSubmatch(pe.name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (pe *publishEvent) parse() bool {
	re := regexp.MustCompile(`-\d+-`)
	if re == nil {
		return false
	}

	matchList := re.Split(pe.name, -1)
	pe.service = matchList[0]

	afterList := strings.Split(matchList[1], "-")
	pe.phase = afterList[0]

	result := re.FindAllStringSubmatch(pe.name, -1)
	matchResult := result[0][0]
	serviceIDStr := strings.Trim(matchResult, "-")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse service id convert to int64 error: %s", err)
		return false
	}
	pe.serviceID = serviceID
	return true
}

func (pe *publishEvent) operate() bool {
	pipeline, err := objects.GetServicePipeline(pe.serviceID)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("query pipeline by service error: %s", err)
		return false
	}
	pipelineID := pipeline.Pipeline.ID

	// 判断该上线流程是否完成
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		log.Info("check deploy is finished, stop update event log")
		return false
	}

	kind := model.PHASE_DEPLOY
	if g.Ini(pipeline.Pipeline.Status, []int{model.PLRollbacking, model.PLRollbackSuccess, model.PLRollbackFailed}) {
		kind = model.PHASE_ROLLBACK
	}

	log.Infof("get pipeline: %d kind: %s phase: %s", pipelineID, kind, pe.phase)

	info := pe.fields[0]
	operTime := info.Time

	msg := fmt.Sprintf("[%s] %v\n%s", operTime, pe.name, pe.message)
	err = objects.RealtimeLog(pipelineID, kind, pe.phase, msg)
	if !errors.Is(err, objects.NotFound) && err != nil {
		log.Errorf("write event log to db error: %s", err)
		return false
	}
	return true
}
