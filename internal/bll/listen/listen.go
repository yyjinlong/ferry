// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

import (
	"io/ioutil"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"ferry/internal/bll/listen/event"
	"ferry/pkg/g"
	"ferry/pkg/log"
)

func getClientset() *kubernetes.Clientset {
	log.InitFields(log.Fields{"logid": g.UniqueID(), "type": "listen"})

	config, err := ioutil.ReadFile(g.Config().K8S.Kubeconfig)
	if err != nil {
		log.Panicf("read kubeconfig file error: %s", err)
	}

	kubeconfig, err := clientcmd.RESTConfigFromKubeConfig(config)
	if err != nil {
		log.Panicf("get rest config from kubeconfig error: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Panicf("get clientset from kubeconfig error: %s", err)
	}
	return clientset
}

func DeploymentFinishEvent() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(getClientset(), 0)
	deploymentInformer := sharedInformer.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.HandleDeploymentCapturer(obj, event.Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.HandleDeploymentCapturer(newObj, event.Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	deploymentInformer.Run(stopCh)
}

func EndpointFinishEvent() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(getClientset(), 0)
	endpointInformer := sharedInformer.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.HandleEndpointCapturer(obj, event.Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.HandleEndpointCapturer(newObj, event.Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	endpointInformer.Run(stopCh)
}

func PublishLogEvent() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(getClientset(), 0)
	eventInformer := sharedInformer.Core().V1().Events().Informer()
	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.HandleLogCapturer(obj, event.Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.HandleLogCapturer(newObj, event.Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	eventInformer.Run(stopCh)
}
