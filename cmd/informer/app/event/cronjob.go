// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package event

import (
	"regexp"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"nautilus/golib/log"
)

func HandleJobCapturer(obj interface{}, mode string) {
	var (
		data          = obj.(*batchv1.Job)
		name          = data.ObjectMeta.Name
		successPodNum = data.Status.Succeeded
		failPodNum    = data.Status.Failed
		beginTime     = data.Status.StartTime
		finishTime    = data.Status.CompletionTime
	)

	log.InitFields(log.Fields{
		"mode":    mode,
		"job":     name,
		"version": data.ObjectMeta.ResourceVersion,
	})

	handleEvent(&JobCapturer{
		mode:          mode,
		name:          name,
		successPodNum: successPodNum,
		failPodNum:    failPodNum,
		beginTime:     beginTime,
		finishTime:    finishTime,
	})
}

type JobCapturer struct {
	mode          string
	name          string       // job名称
	successPodNum int32        // 运行成功的pod数量
	failPodNum    int32        // 运行失败的pod数量
	beginTime     *metav1.Time // job开始时间
	finishTime    *metav1.Time // job结束时间
	jobResult     int          // job运行结果 0 运行中 1 运行成功 2 运行失败
	jobID         int64        // job id
	service       string       // job对应的服务
}

func (jc *JobCapturer) valid() bool {
	// NOTE: 检查是否是业务的cronjob
	reg := regexp.MustCompile(`[\w+-]-cronjob-\d+-`)
	if reg == nil {
		return false
	}
	result := reg.FindAllStringSubmatch(jc.name, -1)
	if len(result) == 0 {
		return false
	}
	return true
}

func (jc *JobCapturer) ready() bool {
	if jc.successPodNum >= 1 {
		jc.jobResult = 1
	} else if jc.failPodNum >= 1 {
		jc.jobResult = 2
	} else {
		jc.jobResult = 0
	}
	return true
}

func (jc *JobCapturer) parse() bool {
	reg := regexp.MustCompile(`-\d+-`)

	// 获取服务ID
	result := reg.FindAllStringSubmatch(jc.name, -1)
	matchResult := result[0][0]
	jobIDStr := strings.Trim(matchResult, "-")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		log.Errorf("parse job id convert to int64 error: %s", err)
		return false
	}
	jc.jobID = jobID

	// 获取服务名
	matchList := reg.Split(jc.name, -1)
	before := matchList[0]
	beforeList := strings.Split(before, "-cronjob")
	jc.service = beforeList[0]
	return true
}

func (jc *JobCapturer) operate() bool {
	if jc.jobResult == 1 {
		log.Infof("service: %s job: %d begin: %v end: %v running success", jc.service, jc.jobID, jc.beginTime, jc.finishTime)
	} else if jc.jobResult == 2 {
		log.Infof("service: %s job: %d begin: %v end: %v running failed", jc.service, jc.jobID, jc.beginTime, jc.finishTime)
	}
	return true
}
