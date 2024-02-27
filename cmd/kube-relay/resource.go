package main

import (
	"time"

	"github.com/anhk/kube-relay/pkg/log"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type ResourceHandler struct {
	GVR    schema.GroupVersionResource
	Lister cache.GenericLister
}

func (res *ResourceHandler) WatchFunc(ctx *gin.Context) {

}

func (res *ResourceHandler) AddFunc(obj any) {
	o := objectToUnstructured(obj)
	// log.Debug("[%v] GetAPIVersion: %v", res.gvr.Resource, o.GetAPIVersion())

	// kind := o.GetObjectKind()
	// log.Debug("[%v] GroupVersionKind: %v", res.gvr.Resource, kind.GroupVersionKind())
	// log.Debug("GetResourceVersion: %v", o.GetResourceVersion())

	log.Debug("[%v] add [%v] %v/%v", res.GVR.Resource, o.GetKind(), o.GetNamespace(), o.GetName())
}

func (res *ResourceHandler) UpdateFunc(oldObj, newObj any) {
	log.Debug("update")
}

func (res *ResourceHandler) DeleteFunc(obj any) {
	log.Debug("delete")
}

func (res *ResourceHandler) RunWithDynamicClient(dynamicClient dynamic.Interface) cache.InformerSynced {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 30*time.Minute, "", nil)
	informer := factory.ForResource(res.GVR)
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: res.AddFunc, UpdateFunc: res.UpdateFunc, DeleteFunc: res.DeleteFunc,
	})
	go informer.Informer().Run(wait.NeverStop)
	res.Lister = informer.Lister()
	return informer.Informer().HasSynced
}

func NewResourceHandler(gvr schema.GroupVersionResource) *ResourceHandler {
	log.Info("resource=%v, group=%v, version=%v", gvr.Resource, gvr.Group, gvr.Version)
	return &ResourceHandler{GVR: gvr}
}
