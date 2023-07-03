// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"fmt"
	"regexp"
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

	serviceName, phase := r.parseInfo(name)
	if !cm.In(phase, []string{model.PHASE_SANDBOX, model.PHASE_ONLINE}) {
		return nil
	}

	pipeline, err := model.GetServicePipeline(serviceName)
	if err != nil {
		log.Errorf("[log] query pipeline by service error: %s", err)
		return err
	}

	// 判断该上线流程是否完成
	if cm.Ini(pipeline.Status, []int{model.PLSuccess, model.PLRollbackSuccess}) {
		return nil
	}
	pipelineID := pipeline.ID

	kind := model.KIND_DEPLOY
	if pipeline.Status == model.PLRollbacking {
		kind = model.KIND_ROLLBACK
	}

	info := fields[0]
	operTime := info.Time.Format("15:04:05")

	msg := fmt.Sprintf("[%s] %v\n%s", operTime, serviceName, message)
	model.RealtimeLog(pipelineID, kind, phase, msg)
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

func (r *LogResource) parseInfo(name string) (string, string) {
	re := regexp.MustCompile(`-\d+-`)
	matchList := re.Split(name, -1)
	service := matchList[0]

	afterList := strings.Split(matchList[1], "-")
	phase := afterList[0]
	return service, phase
}
