// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package trace

import (
	"io/ioutil"
	"time"

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

func ListenEvent() {
	log.InitFields(log.Fields{"logid": g.UniqueID(), "type": "trace"})

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

	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, time.Minute)

	deploymentInformer := sharedInformer.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleDeployment(obj, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			handleDeployment(newObj, Update)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	deploymentInformer.Run(stopCh)

	endpointInformer := sharedInformer.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleEndpoint(nil, obj, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			handleEndpoint(oldObj, newObj, Create)
		},
		DeleteFunc: func(obj interface{}) {
			handleEndpoint(nil, obj, Create)
		},
	})
	endpointInformer.Run(stopCh)
}
