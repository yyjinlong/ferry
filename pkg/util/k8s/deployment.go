package k8s

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment interface {
	GetDeployment(namespace, name string) (*appsv1.Deployment, error)
	GetDeploymentPods(namespace, name string) (*corev1.PodList, error)
	CreateDeployment(namespace string, deployment *appsv1.Deployment) error
	UpdateDeployment(namespace string, deployment *appsv1.Deployment) error
	CreateOrUpdateDeployment(namespace string, deployment *appsv1.Deployment) error
	DeleteDeployment(namespace, name string) error
	ListDeployments(namespace string) (*appsv1.DeploymentList, error)
	Scale(namespace, name string, replicas int32) error
}

type DeploymentResource struct {
	clientset *kubernetes.Clientset
}

func NewDeploymentResource(clientset *kubernetes.Clientset) *DeploymentResource {
	return &DeploymentResource{
		clientset: clientset,
	}
}

func (r *DeploymentResource) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := r.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (r *DeploymentResource) GetDeploymentPods(namespace, name string) (*corev1.PodList, error) {
	deployment, err := r.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	labels := []string{}
	for k, v := range deployment.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return r.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
}

func (r *DeploymentResource) CreateDeployment(namespace string, deployment *appsv1.Deployment) error {
	if _, err := r.clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{}); err != nil {
		return err
	}
	log.Debugf("namespace: %s deployment: %s created", namespace, deployment.ObjectMeta.Name)
	return nil
}

func (r *DeploymentResource) UpdateDeployment(namespace string, deployment *appsv1.Deployment) error {
	if _, err := r.clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}
	log.Debugf("namespace: %s deployment: %s updated", namespace, deployment.ObjectMeta.Name)
	return nil
}
func (r *DeploymentResource) CreateOrUpdateDeployment(namespace string, deployment *appsv1.Deployment) error {
	storedDeployment, err := r.GetDeployment(namespace, deployment.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return r.CreateDeployment(namespace, deployment)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version.
	deployment.ResourceVersion = storedDeployment.ResourceVersion
	return r.UpdateDeployment(namespace, deployment)
}

func (r *DeploymentResource) DeleteDeployment(namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return r.clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func (r *DeploymentResource) ListDeployments(namespace string) (*appsv1.DeploymentList, error) {
	return r.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
}

func (r *DeploymentResource) Scale(namespace, name string, replicas int32) error {
	deployment, err := r.GetDeployment(namespace, name)
	if err != nil {
		return err
	}
	deployment.Spec.Replicas = &replicas
	return r.UpdateDeployment(namespace, deployment)
}
