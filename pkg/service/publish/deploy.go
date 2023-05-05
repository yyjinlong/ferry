// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"encoding/json"
	"fmt"

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

func NewDeploy() *Deploy {
	return &Deploy{}
}

type Deploy struct{}

func (d *Deploy) Handle(pid int64, phase, username string) error {
	pipeline, err := model.GetPipeline(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_QUERY_ERROR, pid, err)
	}

	if err := d.checkStatus(pipeline.Status); err != nil {
		return err
	}

	svc, err := model.GetServiceByID(pipeline.ServiceID)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	ns, err := model.GetNamespaceByID(svc.NamespaceID)
	if err != nil {
		return fmt.Errorf(config.DB_QUERY_NAMESPACE_ERROR, err)
	}

	var (
		serviceID      = pipeline.ServiceID
		serviceName    = svc.Name
		deployGroup    = svc.DeployGroup
		replicas       = int32(svc.Replicas)
		graceTime      = int64(svc.ReserveTime)
		namespace      = ns.Name
		deploymentName = k8s.GetDeploymentName(serviceName, serviceID, phase, deployGroup)
		configMapName  = k8s.GetConfigmapName(serviceName)
		labels         = d.generateLabels(serviceName, phase, deploymentName)
	)

	if phase == model.PHASE_SANDBOX {
		// 沙盒阶段默认返回1个副本
		replicas = 1
	}
	log.Infof("publish get deployment name: %s group: %s replicas: %d", deploymentName, deployGroup, replicas)

	imageInfo, err := model.FindImageInfo(pid)
	if err != nil {
		return fmt.Errorf(config.DB_PIPELINE_UPDATE_ERROR, err)
	}

	if len(imageInfo) == 0 {
		return fmt.Errorf(config.PUB_FETCH_IMAGE_INFO_ERROR)
	}
	log.Infof("publish get deployment image info: %s", imageInfo)

	volumes, err := d.generateVolumes(svc.Volume)
	if err != nil {
		return fmt.Errorf(config.PUB_CREATE_VOLUMES_ERROR, err)
	}

	volumeMounts, err := d.generateVolumeMounts(svc.Volume)
	if err != nil {
		return fmt.Errorf(config.PUB_CREATE_VOLUME_MOUNT_ERROR, err)
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
					Affinity:                      d.generateAffinity(deploymentName),
					NodeSelector:                  d.generateNodeSelector(),
					SecurityContext:               d.generatePodSecurity(),
					DNSPolicy:                     corev1.DNSClusterFirst,
					DNSConfig:                     d.generatePodDNSConfig(),
					ImagePullSecrets:              d.generateImagePullSecret(),
					HostAliases:                   d.generateHostAlias(),
					Volumes:                       volumes,
					TerminationGracePeriodSeconds: &graceTime,
					Containers: []corev1.Container{
						{
							Name:            serviceName,
							Image:           fmt.Sprintf("%s:%s", imageInfo["image_url"], imageInfo["image_tag"]),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             d.generateEnvs(namespace, serviceName, phase),
							EnvFrom:         d.generateEnvFroms(configMapName),
							SecurityContext: d.generateContainerSecurity(),
							Resources:       d.generateResources(svc.QuotaCPU, svc.QuotaMaxCPU, svc.QuotaMem, svc.QuotaMaxMem),
							VolumeMounts:    volumeMounts,
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

func (d *Deploy) checkStatus(status int) error {
	statusList := []int{
		model.PLSuccess,
		model.PLFailed,
		model.PLRollbackSuccess,
		model.PLRollbackFailed,
		model.PLTerminate,
	}
	if cm.Ini(status, statusList) {
		return fmt.Errorf(config.PUB_DEPLOY_FINISHED)
	}
	return nil
}

func (d *Deploy) generateLabels(service, phase, deploymentName string) map[string]string {
	return map[string]string{
		"service": service,
		"phase":   phase,
		"appid":   deploymentName,
	}
}

// 同一deployment下的pod散列在不同node上
func (d *Deploy) generateAffinity(deploymentName string) *corev1.Affinity {
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
									Key:      "service",
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

func (d *Deploy) generateNodeSelector() map[string]string {
	return map[string]string{
		"aggregate": "default",
	}
}

func (d *Deploy) generatePodSecurity() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{}
}

func (d *Deploy) generatePodDNSConfig() *corev1.PodDNSConfig {
	return &corev1.PodDNSConfig{
		Nameservers: []string{"114.114.114.114"},
	}
}

func (d *Deploy) generateImagePullSecret() []corev1.LocalObjectReference {
	return []corev1.LocalObjectReference{
		{
			Name: config.Config().K8S.ImageKey,
		},
	}
}

func (d *Deploy) generateHostAlias() []corev1.HostAlias {
	return []corev1.HostAlias{
		{
			IP:        "127.0.0.1",
			Hostnames: []string{"localhost.localdomain"},
		},
	}
}

func (d *Deploy) generateVolumes(volumeConf string) ([]corev1.Volume, error) {
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

func (d *Deploy) generateEnvs(namespace, service, stage string) []corev1.EnvVar {
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

func (d *Deploy) generateEnvFroms(configMapName string) []corev1.EnvFromSource {
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

func (d *Deploy) generateContainerSecurity() *corev1.SecurityContext {
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

func (d *Deploy) generateResources(minCPU, maxCPU, minMem, maxMem int) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", minCPU)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", minMem)),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", maxCPU)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", maxMem)),
		},
	}
}

func (d *Deploy) generateVolumeMounts(volumeConf string) ([]corev1.VolumeMount, error) {
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
