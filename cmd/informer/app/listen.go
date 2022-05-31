// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package app

import (
	"io/ioutil"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"nautilus/cmd/informer/app/event"
	"nautilus/golib/log"
	"nautilus/pkg/config"
)

func getClusterConfig(cluster string) string {
	switch cluster {
	case config.HP:
		return config.Config().K8S.HPConfig
	case config.XQ:
		return config.Config().K8S.XQConfig
	}
	return ""
}

func GetClientset(cluster string) *kubernetes.Clientset {
	clusterConfig := getClusterConfig(cluster)
	config, err := ioutil.ReadFile(clusterConfig)
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

func DeploymentFinishEvent(clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
	deploymentInformer := sharedInformer.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.HandleDeploymentCapturer(obj, event.Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDep := oldObj.(*appsv1.Deployment)
			newDep := newObj.(*appsv1.Deployment)
			if oldDep.ObjectMeta.ResourceVersion == newDep.ObjectMeta.ResourceVersion {
				return
			}
			event.HandleDeploymentCapturer(newObj, event.Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	deploymentInformer.Run(stopCh)
}

func PublishLogEvent(clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
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

func EndpointReadyEvent(clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, time.Minute)
	endpointInformer := sharedInformer.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event.HandleEndpointCapturer(obj, event.Create, clientset)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldEnd := oldObj.(*corev1.Endpoints)
			newEnd := newObj.(*corev1.Endpoints)
			if oldEnd.ObjectMeta.ResourceVersion == newEnd.ObjectMeta.ResourceVersion {
				return
			}
			event.HandleEndpointCapturer(newObj, event.Update, clientset)
		},
		DeleteFunc: func(obj interface{}) {
			event.HandleEndpointCapturer(obj, event.Delete, clientset)
		},
	})
	endpointInformer.Run(stopCh)
}
