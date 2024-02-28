package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/anhk/kube-relay/pkg/k8s"
	"github.com/anhk/kube-relay/pkg/log"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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

	// TODO: FIFO 队列
}

type ListWrapper struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta  `json:"metadata"`
	Items           []runtime.Object `json:"items"`
}

func (res *ResourceHandler) ListFunc(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	log.Debug("HTTP: [%v] %v/%v", res.GVR, namespace, name)

	list, err := res.Lister.List(labels.Everything())
	if err != nil {
		ctx.AbortWithError(502, err)
		return
	}

	lw := &ListWrapper{}
	lw.APIVersion = res.GVR.Version
	lw.Kind = fmt.Sprintf("%vList", res.apiRes.Kind)
	lw.Items = list
	lw.Metadata.ResourceVersion = "1"

	ctx.JSON(200, lw)
}

func (res *ResourceHandler) WatchFunc(ctx *gin.Context) {
	ctx.Header("content-type", "application/json")

	watch := ctx.Query("watch")

	if watch != "1" && watch != "true" {
		res.ListFunc(ctx)
		return
	}
	resourceVersion := ctx.Query("resourceVersion")
	log.Debug("watch: %v, resourceVersion: %v", watch, resourceVersion)
	// TODO: resourceVersion如果不存在，返回410 Gone

	list, err := res.Lister.List(labels.Everything())
	if err != nil {
		ctx.AbortWithError(502, err)
		return
	}

	for _, obj := range list {
		event := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: obj}}
		data, _ := json.Marshal(event)
		ctx.Writer.Write(data)
	}

	ctx.Stream(func(w io.Writer) bool {
		time.Sleep(time.Second)
		return true
	})
}

func (res *ResourceHandler) AddFunc(obj any) {
	o := k8s.ObjectToUnstructured(obj)
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
