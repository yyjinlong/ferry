package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Service interface {
	GetService(namespace, name string) (*corev1.Service, error)
	CreateService(namespace string, service *corev1.Service) error
	CreateIfNotExistsService(namespace string, service *corev1.Service) error
	UpdateService(namespace string, service *corev1.Service) error
	CreateOrUpdateService(namespace string, service *corev1.Service) error
	DeleteService(namespace, name string) error
	ListServices(namespace string) (*corev1.ServiceList, error)
}

type ServiceResource struct {
	clientset *kubernetes.Clientset
}

func NewServiceResource(clientset *kubernetes.Clientset) *ServiceResource {
	return &ServiceResource{
		clientset: clientset,
	}
}

func (s *ServiceResource) GetService(namespace, name string) (*corev1.Service, error) {
	return s.clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (s *ServiceResource) CreateService(namespace string, service *corev1.Service) error {
	_, err := s.clientset.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	return err
}

func (s *ServiceResource) CreateIfNotExistsService(namespace string, service *corev1.Service) error {
	if _, err := s.GetService(namespace, service.Name); err != nil {
		if errors.IsNotFound(err) {
			return s.CreateService(namespace, service)
		}
		return err
	}
	return nil
}

func (s *ServiceResource) UpdateService(namespace string, service *corev1.Service) error {
	_, err := s.clientset.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	return err
}

func (s *ServiceResource) CreateOrUpdateService(namespace string, service *corev1.Service) error {
	storedService, err := s.GetService(namespace, service.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return s.CreateService(namespace, service)
		}
		return err
	}

	service.ResourceVersion = storedService.ResourceVersion
	return s.UpdateService(namespace, service)
}

func (s *ServiceResource) DeleteService(namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return s.clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func (s *ServiceResource) ListServices(namespace string) (*corev1.ServiceList, error) {
	return s.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
}
