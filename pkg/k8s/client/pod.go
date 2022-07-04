package client

import (
	"bytes"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPods(clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func GetPodName(pods *corev1.PodList, ip string) (string, error) {
	for _, obj := range pods.Items {
		podName := obj.ObjectMeta.Name
		podIP := obj.Status.PodIP
		if podIP == ip {
			return podName, nil
		}
	}
	return "", nil
}

func GetPodLogs(clientset *kubernetes.Clientset, namespace, podname string) (string, error) {
	var line int64 = 100
	request := clientset.CoreV1().Pods(namespace).GetLogs(podname, &corev1.PodLogOptions{TailLines: &line})
	podLogs, err := request.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	logs := buf.String()
	return logs, nil
}
