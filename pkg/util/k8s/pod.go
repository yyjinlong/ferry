package k8s

import (
	"bytes"
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Pod interface {
	GetPod(namespace, name string) (*corev1.Pod, error)
	ListPods(namespace string) (*corev1.PodList, error)
	GetPodLogs(namespace, name string) (string, error)
}

type PodResource struct {
	clientset *kubernetes.Clientset
}

func NewPodResouce(clientset *kubernetes.Clientset) *PodResource {
	return &PodResource{
		clientset: clientset,
	}
}

func (p *PodResource) GetPod(namespace, name string) (*corev1.Pod, error) {
	pod, err := p.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (p *PodResource) ListPods(namespace string) (*corev1.PodList, error) {
	return p.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
}

func (p *PodResource) GetPodLogs(namespace, name string) (string, error) {
	var line int64 = 100
	request := p.clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{TailLines: &line})
	logs, err := request.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, logs); err != nil {
		return "", err
	}
	return buf.String(), nil
}
