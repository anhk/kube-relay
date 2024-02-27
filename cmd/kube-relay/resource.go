package main

import (
	"time"

	"github.com/anhk/kube-relay/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type ResourceHandler struct {
	gvr    schema.GroupVersionResource
	client dynamic.Interface
}

func (res *ResourceHandler) AddFunc(obj any) {
	o := objectToUnstructured(obj)
	// log.Debug("[%v] GetAPIVersion: %v", res.gvr.Resource, o.GetAPIVersion())

	// kind := o.GetObjectKind()
	// log.Debug("[%v] GroupVersionKind: %v", res.gvr.Resource, kind.GroupVersionKind())
	// log.Debug("GetResourceVersion: %v", o.GetResourceVersion())

	log.Debug("[%v] add [%v] %v/%v", res.gvr.Resource, o.GetKind(), o.GetNamespace(), o.GetName())
}

func (res *ResourceHandler) UpdateFunc(oldObj, newObj any) {
	log.Debug("update")
}

func (res *ResourceHandler) DeleteFunc(obj any) {
	log.Debug("delete")
}

func (res *ResourceHandler) Run() cache.InformerSynced {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(res.client, 30*time.Minute, "", nil)
	informer := factory.ForResource(res.gvr)
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: res.AddFunc, UpdateFunc: res.UpdateFunc, DeleteFunc: res.DeleteFunc,
	})
	go informer.Informer().Run(wait.NeverStop)
	return informer.Informer().HasSynced
}

func NewResourceHandlerByDynamicClient(gvr schema.GroupVersionResource,
	dynamicClient dynamic.Interface) *ResourceHandler {
	log.Info("resource=%v, group=%v, version=%v", gvr.Resource, gvr.Group, gvr.Version)
	return &ResourceHandler{gvr: gvr, client: dynamicClient}
}
