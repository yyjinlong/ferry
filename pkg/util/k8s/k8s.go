package k8s

import (
	"fmt"

	"nautilus/pkg/config"
	"nautilus/pkg/model"
)

type Resource interface {
	Deployment
	Service
	ConfigMap
	Pod
	CronJob
	Secret
}

type resource struct {
	Deployment
	Service
	ConfigMap
	Pod
	CronJob
	Secret
}

func New(namespace string) (Resource, error) {
	cluster, err := model.GetClusterByNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf(config.DB_QUERY_CLUSTER_ERROR, err)
	}

	clientset, err := config.GetClientset(cluster)
	if err != nil {
		return nil, fmt.Errorf(config.PUB_GET_CLIENTSET_ERROR, err)
	}

	return &resource{
		Deployment: NewDeploymentResource(clientset),
		Service:    NewServiceResource(clientset),
		ConfigMap:  NewConfigMapResource(clientset),
		Pod:        NewPodResouce(clientset),
		CronJob:    NewCronJobResource(clientset),
		Secret:     NewSecretResource(clientset),
	}, nil
}
