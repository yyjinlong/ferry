// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"ferry/ops/log"
)

func NewEvent(config []byte) *Event {
	kubeconfig, err := clientcmd.RESTConfigFromKubeConfig(config)
	if err != nil {
		log.Panicf("get rest config from kubeconfig error: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Panicf("get clientset from kubeconfig error: %s", err)
	}

	sharedInformer := informers.NewSharedInformerFactory(clientset, time.Minute)
	return &Event{
		sharedInformer: sharedInformer,
	}
}

type Event struct {
	sharedInformer informers.SharedInformerFactory
}

func (e *Event) Deployment() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	deploymentInformer := e.sharedInformer.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    DeploymentAdd,
		UpdateFunc: DeploymentUpdate,
		DeleteFunc: DeploymentDelete,
	})
	deploymentInformer.Run(stopCh)
}

func (e *Event) Endpoint() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	endpointInformer := e.sharedInformer.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    EndpointAdd,
		UpdateFunc: EndpointUpdate,
		DeleteFunc: EndpointDelete,
	})
	endpointInformer.Run(stopCh)
}
