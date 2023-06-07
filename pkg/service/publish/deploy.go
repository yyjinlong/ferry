// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/cm"
	"nautilus/pkg/util/k8s"
)

const (
	CodeMountPoint = "www"
	CodeMountPath  = "/home/tong/www"
	LogMountPoint  = "log"
	LogMountPath   = "/home/tong/www/log"
)

func NewDeploy(pid int64, phase, username string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	statusList := []int{
		model.PLSuccess,
		model.PLFailed,
		model.PLRollbackSuccess,
		model.PLRollbackFailed,
		model.PLTerminate,
	}
	if cm.Ini(pipeline.Status, statusList) {
		return fmt.Errorf(config.PUB_DEPLOY_FINISHED)
	}

	svc, err := model.GetServiceInfo(pipeline.Service)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		serviceID      = svc.ID
		serviceName    = svc.Name
		namespace      = svc.Namespace
		serviceImage   = svc.ImageAddr
		deployGroup    = svc.DeployGroup
		replicas       = int32(svc.Replicas)
		graceTime      = int64(svc.ReserveTime)
		deploymentName = k8s.GetDeploymentName(serviceName, serviceID, phase, deployGroup)
		configMapName  = k8s.GetConfigmapName(serviceName)
		labels         = generateLabels(serviceName, phase, deploymentName)
	)

	if phase == model.PHASE_SANDBOX {
		// 沙盒阶段默认返回1个副本
		replicas = 1
	}
	log.Infof("publish get deployment name: %s group: %s replicas: %d", deploymentName, deployGroup, replicas)

	initContainers, err := generateInitContainers(pid)
	if err != nil {
		return fmt.Errorf(config.PUB_INIT_CONTINAER_ERROR, err)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "100%",
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity:                      generateAffinity(deploymentName),
					NodeSelector:                  generateNodeSelector("default"),
					SecurityContext:               generatePodSecurity(),
					DNSPolicy:                     corev1.DNSClusterFirst,
					DNSConfig:                     generatePodDNSConfig(),
					ImagePullSecrets:              generateImagePullSecret(),
					Volumes:                       generateVolumes(namespace, serviceName, "business", pid),
					TerminationGracePeriodSeconds: &graceTime,
					InitContainers:                initContainers,
					Containers: []corev1.Container{
						{
							Name:            serviceName,
							Image:           serviceImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             generateEnvs(namespace, serviceName, phase),
							EnvFrom:         generateEnvFroms(configMapName),
							SecurityContext: generateContainerSecurity(),
							Resources:       generateResources(svc.QuotaCPU, svc.QuotaMaxCPU, svc.QuotaMem, svc.QuotaMaxMem),
							VolumeMounts:    generateMainVolumeMounts(),
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "sleep 30"},
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 5,
								TimeoutSeconds:      10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    10,
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"/home/tong/opbin/readiness-probe.sh",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: 5,
								TimeoutSeconds:      5,
								PeriodSeconds:       60,
								SuccessThreshold:    1,
								FailureThreshold:    3,
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"/home/tong/opbin/liveness-prob.sh",
										},
									},
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
		return err
	}
	if err := resource.CreateOrUpdateDeployment(namespace, dep); err != nil {
		return fmt.Errorf(config.PUB_K8S_DEPLOYMENT_EXEC_FAILED, err)
	}
	log.Infof("publish deployment: %s to k8s success", deploymentName)

	if err := model.CreatePhase(pid, model.PHASE_DEPLOY, phase, model.PHProcess); err != nil {
		return fmt.Errorf(config.PUB_RECORD_DEPLOYMENT_TO_DB_ERROR, err)
	}
	log.Infof("record deployment: %s to db success", deploymentName)
	return nil
}

func generateLabels(service, phase, deploymentName string) map[string]string {
	return map[string]string{
		"service": service,
		"phase":   phase,
		"appid":   deploymentName,
	}
}

// 同一deployment下的pod散列在不同node上
func generateAffinity(deploymentName string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "appid",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{deploymentName},
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateNodeSelector(zone string) map[string]string {
	return map[string]string{
		"aggregate": zone,
	}
}

func generatePodSecurity() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{}
}

func generatePodDNSConfig() *corev1.PodDNSConfig {
	return &corev1.PodDNSConfig{
		Nameservers: []string{"114.114.114.114"},
	}
}

func generateImagePullSecret() []corev1.LocalObjectReference {
	return []corev1.LocalObjectReference{
		{
			Name: config.Config().K8S.ImageKey,
		},
	}
}

func generateVolumes(namespace, service, category string, pid int64) []corev1.Volume {
	// NOTE: 在宿主机上创建本地存储卷, 目前只支持hostPath-DirectoryOrCreate类型.
	var (
		pathType     = corev1.HostPathDirectoryOrCreate
		codeHostPath = fmt.Sprintf("/home/www/%s/%s/%d", category, service, pid)
		logHostPath  = fmt.Sprintf("/home/log/%s/%s/%s", category, namespace, service)
	)

	return []corev1.Volume{
		{
			Name: CodeMountPoint,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &pathType,
					Path: codeHostPath,
				},
			},
		},
		{
			Name: LogMountPoint,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &pathType,
					Path: logHostPath,
				},
			},
		},
	}
}

func generateInitContainers(pid int64) ([]corev1.Container, error) {
	var containers []corev1.Container

	images, err := model.FindImages(pid)
	if err != nil {
		log.Errorf("query pipeline: %d images error: %+v", pid, err)
		return nil, err
	}

	if len(images) == 0 {
		log.Errorf("query pipeline: %d images empty", pid)
		return nil, err
	}
	log.Infof("publish get deployment images: %s", images)

	for _, item := range images {
		codeModule := strings.Replace(item.CodeModule, "_", "-", -1)
		containers = append(containers, getInitContainer(codeModule, item.ImageURL, item.ImageTag))
	}
	return containers, nil
}

func getInitContainer(module, imageURL, imageTag string) corev1.Container {
	// 约定代码目录: /home/tong/www
	// 约定日志目录: /home/tong/www/log
	var (
		lockFile = fmt.Sprintf("/home/tong/www/%s_done", module)
		cmd      = fmt.Sprintf("cp -rfp /code/* /home/tong/www; chown tong:tong /home/tong/www -R; touch %s", lockFile)
		safeCmd  = fmt.Sprintf(`if [ ! -f %s ]; then %s; fi`, lockFile, cmd)
	)

	return corev1.Container{
		Name:            fmt.Sprintf("%s-code", module),
		Image:           fmt.Sprintf("%s:%s", imageURL, imageTag),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		VolumeMounts: generateInitVolumeMounts(),
		Command:      []string{"/bin/sh", "-c", safeCmd},
	}
}

func generateInitVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      CodeMountPoint,
			MountPath: CodeMountPath,
		},
	}
}

func generateEnvs(namespace, service, stage string) []corev1.EnvVar {
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
			Value: stage,
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

func generateEnvFroms(configMapName string) []corev1.EnvFromSource {
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

func generateContainerSecurity() *corev1.SecurityContext {
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

func generateResources(minCPU, maxCPU, minMem, maxMem string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(minCPU),
			corev1.ResourceMemory: resource.MustParse(minMem),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(maxCPU),
			corev1.ResourceMemory: resource.MustParse(maxMem),
		},
	}
}

func generateMainVolumeMounts() []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0)
	initMounts := generateInitVolumeMounts()
	volumeMounts = append(volumeMounts, initMounts...)

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      LogMountPoint,
		MountPath: LogMountPath,
	})
	return volumeMounts
}
