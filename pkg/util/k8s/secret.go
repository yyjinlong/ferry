package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Secret interface {
	GetSecret(namespace, name string) (*corev1.Secret, error)
	CreateSecret(namespace string, secret *corev1.Secret) error
	UpdateSecret(namespace string, secret *corev1.Secret) error
	CreateOrUpdateSecret(namespace string, secret *corev1.Secret) error
}

type SecretResource struct {
	clientset *kubernetes.Clientset
}

func NewSecretResource(clientset *kubernetes.Clientset) *SecretResource {
	return &SecretResource{
		clientset: clientset,
	}
}

func (s *SecretResource) GetSecret(namespace, name string) (*corev1.Secret, error) {
	return s.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (s *SecretResource) CreateSecret(namespace string, secret *corev1.Secret) error {
	_, err := s.clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	return err
}

func (s *SecretResource) UpdateSecret(namespace string, secret *corev1.Secret) error {
	_, err := s.clientset.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}

func (s *SecretResource) CreateOrUpdateSecret(namespace string, secret *corev1.Secret) error {
	storedSecret, err := s.GetSecret(namespace, secret.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return s.CreateSecret(namespace, secret)
		}
	}

	secret.ResourceVersion = storedSecret.ResourceVersion
	return s.UpdateSecret(namespace, secret)
}
