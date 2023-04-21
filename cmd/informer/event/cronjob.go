// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package event

import (
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CronJob interface {
	HandleCronJob(obj interface{}, mode, cluster string) error
}

type CronJobResource struct {
	clientset *kubernetes.Clientset
}

func NewCronJobResource(clientset *kubernetes.Clientset) *CronJobResource {
	return &CronJobResource{
		clientset: clientset,
	}
}

func (r *CronJobResource) HandleCronJob(obj interface{}, mode, cluster string) error {
	var (
		data          = obj.(*batchv1.Job)
		name          = data.ObjectMeta.Name       // job名称
		successPodNum = data.Status.Succeeded      // 运行成功的pod数量
		failPodNum    = data.Status.Failed         // 运行失败的pod数量
		beginTime     = data.Status.StartTime      // job开始时间
		finishTime    = data.Status.CompletionTime // job结束时间
		jobResult     = 0                          // job运行结果 0 运行中 1 运行成功 2 运行失败
	)

	// 检查是否是业务的cronjob
	if !r.filter(name) {
		return nil
	}

	if successPodNum >= 1 {
		jobResult = 1
	} else if failPodNum >= 1 {
		jobResult = 2
	}
	log.Infof("[cronjob] check %s job result: %d on mode: %s", name, jobResult, mode)

	jobID, service, err := r.parseInfo(name)
	if err != nil {
		return err
	}

	if err := r.callback(jobID, service, beginTime, finishTime, jobResult); err != nil {
		return err
	}
	return nil
}

func (r *CronJobResource) filter(name string) bool {
	re := regexp.MustCompile(`[\w+-]-cronjob-\d+-`)
	if re == nil {
		return false
	}
	result := re.FindAllStringSubmatch(name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (r *CronJobResource) parseInfo(name string) (int64, string, error) {
	re := regexp.MustCompile(`-\d+-`)

	// 获取服务ID
	result := re.FindStringSubmatch(name)
	match := strings.Trim(result[0], "-")
	jobID, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		log.Errorf("[cronjob] parse: %s convert to int64 error: %s", name, err)
		return 0, "", err
	}

	// 获取服务名
	matchList := re.Split(name, -1)
	before := matchList[0]
	beforeList := strings.Split(before, "-cronjob")
	service := beforeList[0]
	return jobID, service, nil
}

func (r *CronJobResource) callback(jobID int64, service string, begin, finish *metav1.Time, result int) error {
	log.Infof("service: %s job: %d begin: %v end: %v running result: %d success", service, jobID, begin, finish, result)
	return nil
}
