package main

import (
	"time"

	"github.com/anhk/kube-relay/pkg/k8s"
	"github.com/anhk/kube-relay/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func watchService(option *Option) {
	kubeClient, err := k8s.CreateKubeClient(option.KubeConfig, option.ApiServer)
	PanicIf(err)

	factory := informers.NewSharedInformerFactory(kubeClient, time.Hour*1)
	informer := factory.Core().V1().Services().Informer()
	informer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc := obj.(*v1.Service)
			log.Info("add service: %+v", svc.ObjectMeta)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldSvc, newSvc := oldObj.(*v1.Service), newObj.(*v1.Service)
			log.Info("update service: old:%+v, new: %+v", oldSvc.ObjectMeta, newSvc.ObjectMeta)
		},
		DeleteFunc: func(obj interface{}) {
			svc := obj.(*v1.Service)
			log.Info("delete service: %+v", svc.ObjectMeta)
		},
	})
	factory.Start(wait.NeverStop)

	factory.WaitForCacheSync(wait.NeverStop)
	log.Info("cache ok")

	select {}
}
