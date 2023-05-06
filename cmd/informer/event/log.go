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

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
)

type Log interface {
	HandleLog(obj interface{}, mode, cluster string) error
}

type LogResource struct {
	clientset *kubernetes.Clientset
}

func NewLogResouce(clientset *kubernetes.Clientset) *LogResource {
	return &LogResource{
		clientset: clientset,
	}
}

func (r *LogResource) HandleLog(obj interface{}, mode, cluster string) error {
	var (
		data    = obj.(*corev1.Event)
		name    = data.Name
		message = data.Message
		fields  = data.ObjectMeta.ManagedFields
	)

	if !r.filter(name) {
		return nil
	}

	serviceID, serviceName, phase, err := r.parseInfo(name)
	if err != nil {
		return err
	}
	if !cm.In(phase, []string{model.PHASE_SANDBOX, model.PHASE_ONLINE}) {
		return nil
	}

	pipeline, err := model.GetServicePipeline(serviceID)
	if err != nil {
		log.Errorf("[log] query pipeline by service error: %s", err)
		return err
	}
	pipelineID := pipeline.ID

	// 判断该上线流程是否完成
	if cm.Ini(pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		return nil
	}

	kind := model.PHASE_DEPLOY
	if pipeline.Status == model.PLRollbacking {
		kind = model.PHASE_ROLLBACK
	}

	log.Infof("[log] get pipeline: %d kind: %s phase: %s", pipelineID, kind, phase)

	info := fields[0]
	operTime := info.Time.Format("15:04:05")

	msg := fmt.Sprintf("[%s] %v\n%s", operTime, serviceName, message)
	if err := model.RealtimeLog(pipelineID, kind, phase, msg); errors.Is(err, model.NotFound) {
		return nil
	} else if err != nil {
		log.Errorf("[log] write event log to db error: %s", err)
		return err
	}
	return nil
}

func (r *LogResource) filter(name string) bool {
	re := regexp.MustCompile(`[\w+-]+-\d+-(sandbox|online)-(blue|green)-[\w+-]+`)
	if re == nil {
		return false
	}
	result := re.FindAllStringSubmatch(name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (r *LogResource) parseInfo(name string) (int64, string, string, error) {
	re := regexp.MustCompile(`-\d+-`)
	matchList := re.Split(name, -1)
	service := matchList[0]

	afterList := strings.Split(matchList[1], "-")
	phase := afterList[0]

	result := re.FindStringSubmatch(name)
	match := strings.Trim(result[0], "-")
	serviceID, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		log.Errorf("[cronjob] parse: %s convert to int64 error: %s", name, err)
		return 0, "", "", err
	}
	return serviceID, service, phase, nil
}
