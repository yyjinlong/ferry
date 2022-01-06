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

	"ferry/pkg/g"
	"ferry/pkg/log"
)

const (
	Create = "create"
	Update = "update"
	Delete = "delete"
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
			CheckDeploymentIsFinish(obj, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			CheckDeploymentIsFinish(newObj, Update)
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
			CheckEndpointIsFinish(obj, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			CheckEndpointIsFinish(newObj, Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	endpointInformer.Run(stopCh)
}

func GetProcessEvent() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(getClientset(), 0)
	eventInformer := sharedInformer.Core().V1().Events().Informer()
	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			FetchPublishEvent(obj, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			FetchPublishEvent(newObj, Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	eventInformer.Run(stopCh)
}
