// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"io/ioutil"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"nautilus/cmd/informer/app/event"
	"nautilus/golib/log"
	"nautilus/pkg/config"
)

func getClientset() *kubernetes.Clientset {
	config, err := ioutil.ReadFile(config.Config().K8S.Kubeconfig)
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
