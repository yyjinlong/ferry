// copyright @ 2022 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/k8s"
)

func NewCronjob() *CronJob {
	return &CronJob{}
}

type CronJob struct{}

func (c *CronJob) Handle(namespace, service, command, schedule string) (string, error) {
	crontabID, err := model.CreateCrontab(namespace, service, command, schedule)
	if err != nil {
		return "", fmt.Errorf(config.CRON_WRITE_DB_ERROR, err)
	}

	svc, err := model.GetServiceInfo(service)
	if err != nil {
		return "", fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	pipeline, err := model.GetServicePipeline(svc.ID)
	if err != nil {
		return "", fmt.Errorf(config.DB_PIPELINE_QUERY_FAILED, err)
	}

	imageInfo, err := model.FindImages(pipeline.ID)
	if err != nil {
		return "", fmt.Errorf(config.DB_PIPELINE_UPDATE_ERROR, err)
	}

	if len(imageInfo) == 0 {
		return "", fmt.Errorf("get image info is empty")
	}

	var (
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
	)

	volumes, err := c.generateVolumes(svc.Volume)
	if err != nil {
		return "", fmt.Errorf(config.CRON_CREATE_VOLUMES_ERROR, err)
	}

	volumeMounts, err := c.generateVolumeMounts(svc.Volume)
	if err != nil {
		return "", fmt.Errorf(config.CRON_CREATE_VOLUME_MOUNT_ERROR, err)
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
							Labels: c.generateLabels(service, phase, name),
						},
						Spec: corev1.PodSpec{
							RestartPolicy:                 corev1.RestartPolicyNever,
							NodeSelector:                  c.generateNodeSelector(),
							SecurityContext:               c.generatePodSecurity(),
							DNSPolicy:                     corev1.DNSClusterFirst,
							DNSConfig:                     c.generatePodDNSConfig(),
							ImagePullSecrets:              c.generateImagePullSecret(),
							HostAliases:                   c.generateHostAlias(),
							Volumes:                       volumes,
							TerminationGracePeriodSeconds: &graceTime,
							Containers: []corev1.Container{
								{
									Name:            service,
									Image:           fmt.Sprintf("%s:%s", imageInfo["image_url"], imageInfo["image_tag"]),
									ImagePullPolicy: corev1.PullIfNotPresent,
									Args:            c.generateArgs(command),
									Env:             c.generateEnvs(namespace, service, phase),
									EnvFrom:         c.generateEnvFroms(configMapName),
									SecurityContext: c.generateContainerSecurity(),
									Resources:       c.generateResources(),
									VolumeMounts:    volumeMounts,
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

func (c *CronJob) generateLabels(service, phase, name string) map[string]string {
	return map[string]string{
		"service": service,
		"phase":   phase,
		"appid":   name,
	}
}

func (c *CronJob) generateNodeSelector() map[string]string {
	return map[string]string{
		"batch": "cronjob",
	}
}

func (c *CronJob) generatePodSecurity() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{}
}

func (c *CronJob) generatePodDNSConfig() *corev1.PodDNSConfig {
	return &corev1.PodDNSConfig{
		Nameservers: []string{"114.114.114.114"},
	}
}

func (c *CronJob) generateImagePullSecret() []corev1.LocalObjectReference {
	return []corev1.LocalObjectReference{
		{
			Name: config.Config().K8S.ImageKey,
		},
	}
}

func (c *CronJob) generateHostAlias() []corev1.HostAlias {
	return []corev1.HostAlias{
		{
			IP:        "127.0.0.1",
			Hostnames: []string{"localhost.localdomain"},
		},
	}
}

func (c *CronJob) generateVolumes(volumeConf string) ([]corev1.Volume, error) {
	// NOTE: 在宿主机上创建本地存储卷, 目前只支持hostPath-DirectoryOrCreate类型.
	type volume struct {
		Name      string `json:"name"`
		HostPath  string `json:"host_path"`
		MountPath string `json:"mount_path"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(volumeConf), &volumes); err != nil {
		return nil, err
	}

	pathType := corev1.HostPathDirectoryOrCreate
	podVolumes := make([]corev1.Volume, 0)
	for _, item := range volumes {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: item.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &pathType,
					Path: item.HostPath,
				},
			},
		})
	}
	return podVolumes, nil
}

func (c *CronJob) generateArgs(command string) []string {
	cmd := fmt.Sprintf("su - tong -c \"%s && sleep 10\"", command)
	return []string{"/bin/sh", "-c", cmd}
}

func (c *CronJob) generateEnvs(namespace, service, phase string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: namespace,
		},
		{
			Name:  "SERVICE",
			Value: service,
		},
		{
			Name:  "STAGE",
			Value: phase,
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
	}
}

func (c *CronJob) generateEnvFroms(configMapName string) []corev1.EnvFromSource {
	return []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
}

func (c *CronJob) generateContainerSecurity() *corev1.SecurityContext {
	capabilities := &corev1.Capabilities{
		Add: []corev1.Capability{"SYS_ADMIN", "SYS_PTRACE"},
	}

	privileged := false
	defaultUserAndGroup := int64(0)
	runAsNonRoot := false
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := false

	return &corev1.SecurityContext{
		Capabilities:             capabilities,
		Privileged:               &privileged,
		RunAsUser:                &defaultUserAndGroup,
		RunAsGroup:               &defaultUserAndGroup,
		RunAsNonRoot:             &runAsNonRoot,
		ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
	}
}

func (c *CronJob) generateResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("200Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("4096Mi"),
		},
	}
}

func (c *CronJob) generateVolumeMounts(volumeConf string) ([]corev1.VolumeMount, error) {
	type volume struct {
		Name      string `json:"name"`
		HostPath  string `json:"host_path"`
		MountPath string `json:"mount_path"`
	}

	var volumes []volume
	if err := json.Unmarshal([]byte(volumeConf), &volumes); err != nil {
		return nil, err
	}

	volumeMounts := make([]corev1.VolumeMount, 0)
	for _, item := range volumes {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      item.Name,
			MountPath: item.MountPath,
		})
	}
	return volumeMounts, nil
}

func NewCronJobDelete() *CronJobDelete {
	return &CronJobDelete{}
}

type CronJobDelete struct{}

func (c *CronJobDelete) Handle(namespace, service string, jobID int64) error {
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
