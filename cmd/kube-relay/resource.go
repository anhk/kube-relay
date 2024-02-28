package main

import (
	"fmt"
	"time"

	"github.com/anhk/kube-relay/pkg/log"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ResourceHandler struct {
	GVR    schema.GroupVersionResource
	Lister cache.GenericLister
	apiRes metav1.APIResource
	apiGr  metav1.APIGroup
}

func (res *ResourceHandler) WatchFunc(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	log.Error("HTTP: [%v] %v/%v", res.GVR, namespace, name)

	list, err := res.Lister.List(labels.Everything())
	if err != nil {
		ctx.AbortWithError(502, err)
		return
	}

	ctx.JSON(200, list)
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

func (res *ResourceHandler) GetInfoByKubeClient(kubeClient *kubernetes.Clientset) error {
	var groupVersion = "v1" // CoreAPI
	if res.GVR.Group != "" {
		groupVersion = fmt.Sprintf("%v/%v", res.GVR.Group, res.GVR.Version)
	}

	resourceList, err := kubeClient.DiscoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return err
	}
	for _, v := range resourceList.APIResources {
		if v.Name == res.GVR.Resource {
			res.apiRes = v
			res.apiGr.Name = res.GVR.Group
			res.apiGr.Versions = []metav1.GroupVersionForDiscovery{{GroupVersion: groupVersion, Version: res.GVR.Version}}
			res.apiGr.PreferredVersion = res.apiGr.Versions[0]
			return nil
		}
	}

	return fmt.Errorf("%v not found", res.GVR)
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
