// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	Create = "create"
	Update = "update"
	Delete = "delete"
)

type Event interface {
	Deployment
	Endpoint
	CronJob
	Log
}

type handler struct {
	Deployment
	Endpoint
	CronJob
	Log
}

func NewEvent(clientset *kubernetes.Clientset) Event {
	return handler{
		Deployment: NewDeploymentResource(clientset),
		Endpoint:   NewEndpointResource(clientset),
		CronJob:    NewCronJobResource(clientset),
		Log:        NewLogResouce(clientset),
	}
}

// DeploymentEvent deployment事件
func DeploymentEvent(e Event, cluster string, clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
	deploymentInformer := sharedInformer.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e.HandleDeployment(obj, Create, cluster)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDep := oldObj.(*appsv1.Deployment)
			newDep := newObj.(*appsv1.Deployment)
			if oldDep.ObjectMeta.ResourceVersion == newDep.ObjectMeta.ResourceVersion {
				return
			}
			e.HandleDeployment(newObj, Update, cluster)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	deploymentInformer.Run(stopCh)
}

// EndpointEvent endpoint事件
func EndpointEvent(e Event, cluster string, clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, time.Minute)
	endpointInformer := sharedInformer.Core().V1().Endpoints().Informer()
	endpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e.HandleEndpoint(obj, Create, cluster)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldEnd := oldObj.(*corev1.Endpoints)
			newEnd := newObj.(*corev1.Endpoints)
			if oldEnd.ObjectMeta.ResourceVersion == newEnd.ObjectMeta.ResourceVersion {
				return
			}
			e.HandleEndpoint(newObj, Update, cluster)
		},
		DeleteFunc: func(obj interface{}) {
			e.HandleEndpoint(obj, Delete, cluster)
		},
	})
	endpointInformer.Run(stopCh)
}

// LogEvent 发布日志事件
func LogEvent(e Event, cluster string, clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
	eventInformer := sharedInformer.Core().V1().Events().Informer()
	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e.HandleLog(obj, Create, cluster)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			e.HandleLog(newObj, Update, cluster)
		},
		DeleteFunc: func(obj interface{}) {},
	})
	eventInformer.Run(stopCh)
}

// CronjobEvent cronjob事件
func CronjobEvent(e Event, cluster string, clientset *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformer := informers.NewSharedInformerFactory(clientset, time.Minute)
	cronjobInformer := sharedInformer.Batch().V1().Jobs().Informer()
	cronjobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e.HandleCronJob(obj, Create, cluster)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)
			if oldJob.ObjectMeta.ResourceVersion == newJob.ObjectMeta.ResourceVersion {
				return
			}
			e.HandleCronJob(newObj, Update, cluster)
		},
		DeleteFunc: func(obj interface{}) {
			e.HandleCronJob(obj, Delete, cluster)
		},
	})
	cronjobInformer.Run(stopCh)
}
