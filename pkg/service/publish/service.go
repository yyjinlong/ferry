// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
	"nautilus/pkg/util/k8s"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Handle(serviceName string) error {
	svc, err := model.GetServiceInfo(serviceName)
	if err != nil {
		return fmt.Errorf(config.DB_SERVICE_QUERY_ERROR, err)
	}

	var (
		serviceID     = svc.ID
		namespace     = svc.Namespace
		port          = svc.Port
		containerPort = svc.ContainerPort
	)

	var eg errgroup.Group
	for _, item := range []string{model.PHASE_SANDBOX, model.PHASE_ONLINE} {
		phase := item
		eg.Go(func() error {
			return s.worker(namespace, serviceName, serviceID, phase, k8s.BLUE, port, containerPort)
		})

		eg.Go(func() error {
			return s.worker(namespace, serviceName, serviceID, phase, k8s.GREEN, port, containerPort)
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf(config.SVC_WAIT_ALL_SERVICE_ERROR, err)
	}
	return nil
}

func (s *Service) worker(namespace, serviceName string, serviceID int64, phase, group string, port, containerPort int) error {
	// service名称与deployment名称保持一致
	name := k8s.GetDeploymentName(serviceName, serviceID, phase, group)
	labels := map[string]string{
		"appid": name,
	}

	se := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       name,
					Port:       int32(port),
					TargetPort: intstr.FromInt(containerPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	resource, err := k8s.New(namespace)
	if err != nil {
		return err
	}
	if err := resource.CreateOrUpdateService(namespace, se); err != nil {
		return fmt.Errorf(config.SVC_K8S_SERVICE_EXEC_FAILED, err)
	}
	log.Infof("publish service resource: %s to k8s success", name)
	return nil
}
