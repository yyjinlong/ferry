package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Service interface {
	GetService(namespace, name string) (*corev1.Service, error)
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
	return nil, nil
}
