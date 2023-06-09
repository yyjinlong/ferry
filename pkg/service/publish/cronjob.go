// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/k8s"
)

func NewCronjob(namespace, service, command, schedule string) (string, error) {
	crontabID, err := model.CreateCrontab(namespace, service, command, schedule)
	if err != nil {
		return "", fmt.Errorf(config.CRON_WRITE_DB_ERROR, err)
	}

	svc, err := model.GetServiceInfo(service)
	if err != nil {
		return "", fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	pipeline, err := model.GetServiceLastSuccessPipeline(service)
	if err != nil {
		return "", fmt.Errorf(config.DB_PIPELINE_QUERY_FAILED, err)
	}

	var (
		pid                  = pipeline.ID
		name                 = k8s.GetCronjobName(service, crontabID)
		configMapName        = k8s.GetConfigmapName(service)
		bootDeadline   int64 = 90
		successHistory int32 = 0
		failedHistory  int32 = 0
		suspend              = false
		parallelism    int32 = 1
		completions    int32 = 1
		backoffLimit   int32 = 0
		phase                = "cronjob"
		graceTime            = int64(svc.ReserveTime)
		serviceImage         = svc.ImageAddr
	)

	initContainers, err := generateInitContainers(pid)
	if err != nil {
		return "", fmt.Errorf(config.PUB_INIT_CONTINAER_ERROR, err)
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   schedule,
			ConcurrencyPolicy:          batchv1.ForbidConcurrent, // 类似文件锁
			StartingDeadlineSeconds:    &bootDeadline,            // 开始该任务的截止时间秒数
			SuccessfulJobsHistoryLimit: &successHistory,          // 保留多少已完成的任务数
			FailedJobsHistoryLimit:     &failedHistory,           // 保留多少失败的任务数
			Suspend:                    &suspend,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism:  &parallelism,  // 并发启动pod数目
					Completions:  &completions,  // 至少要完成的pod的数目
					BackoffLimit: &backoffLimit, // job的重试次数
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: generateLabels(service, phase, name),
						},
						Spec: corev1.PodSpec{
							RestartPolicy:                 corev1.RestartPolicyNever,
							NodeSelector:                  generateNodeSelector("cronJob"),
							SecurityContext:               generatePodSecurity(),
							DNSPolicy:                     corev1.DNSClusterFirst,
							DNSConfig:                     generatePodDNSConfig(),
							ImagePullSecrets:              generateImagePullSecret(),
							Volumes:                       generateVolumes(namespace, service, "cronjob", crontabID),
							TerminationGracePeriodSeconds: &graceTime,
							InitContainers:                initContainers,
							Containers: []corev1.Container{
								{
									Name:            service,
									Image:           serviceImage,
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env:             generateEnvs(namespace, service, phase),
									EnvFrom:         generateEnvFroms(configMapName),
									SecurityContext: generateContainerSecurity(),
									Resources:       generateResources("500m", "1000m", "512Mi", "4096Mi"),
									VolumeMounts:    generateMainVolumeMounts(),
									Args:            generateArgs(command),
								},
							},
						},
					},
				},
			},
		},
	}

	resource, err := k8s.New(namespace)
	if err != nil {
		return "", err
	}
	if err := resource.CreateOrUpdateCronJob(namespace, cronJob); err != nil {
		return "", fmt.Errorf(config.CRON_K8S_EXEC_FAILED, err)
	}
	log.Infof("publish cronjob: %s to k8s success", name)
	return name, nil
}

func generateArgs(command string) []string {
	cmd := fmt.Sprintf(`su - tong -c "%s && sleep 5"`, command)
	return []string{"/bin/sh", "-c", cmd}
}

func NewCronJobDelete(namespace, service string, jobID int64) error {
	name := k8s.GetCronjobName(service, jobID)
	log.Infof("delete cronjob name: %s", name)

	resource, err := k8s.New(namespace)
	if err != nil {
		return err
	}

	if err := resource.DeleteCronJob(namespace, name); err != nil {
		log.Errorf("delete cronjob: %s failed: %s", name, err)
		return err
	}
	log.Infof("delete cronjob: %s success", name)
	return nil
}
