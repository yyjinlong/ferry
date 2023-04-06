package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMap interface {
	GetConfigMap(namespace, name string) (*corev1.ConfigMap, error)
	CreateConfigMap(namespace string, configMap *corev1.ConfigMap) error
	UpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error
	CreateOrUpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error
	DeleteConfigMap(namespace, name string) error
	ListConfigMaps(namespace string) (*corev1.ConfigMapList, error)
}

type ConfigMapResource struct {
	clientset *kubernetes.Clientset
}

func NewConfigMapResource(clientset *kubernetes.Clientset) *ConfigMapResource {
	return &ConfigMapResource{
		clientset: clientset,
	}
}

func (c *ConfigMapResource) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c *ConfigMapResource) CreateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	if _, err := c.clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *ConfigMapResource) UpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	if _, err := c.clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *ConfigMapResource) CreateOrUpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	storedConfigMap, err := c.GetConfigMap(namespace, configMap.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return c.CreateConfigMap(namespace, configMap)
		}
	}

	configMap.ResourceVersion = storedConfigMap.ResourceVersion
	return c.UpdateConfigMap(namespace, configMap)
}

func (c *ConfigMapResource) DeleteConfigMap(namespace, name string) error {
	return c.clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *ConfigMapResource) ListConfigMaps(namespace string) (*corev1.ConfigMapList, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
}
