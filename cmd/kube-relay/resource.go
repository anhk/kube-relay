package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

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

	fifo *ResourceFifo
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

	version := res.fifo.Version()
	list, err := res.Lister.List(labels.Everything())
	if err != nil {
		ctx.AbortWithError(502, err)
		return
	}

	lw := &ListWrapper{}
	lw.APIVersion = res.GVR.Version
	lw.Kind = fmt.Sprintf("%vList", res.apiRes.Kind)
	lw.Items = list
	lw.Metadata.ResourceVersion = version

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

	if resourceVersion == "" || resourceVersion == "0" { // 拿全部数据
		resourceVersion = res.fifo.Version()
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
		ctx.Writer.Flush()
	}

	ctx.Stream(func(w io.Writer) bool {
		for {
			if err := res.fifo.Wait(resourceVersion); err != nil {
				ctx.AbortWithError(502, err)
				return false
			}
			list, curVersion, err := res.fifo.Get(resourceVersion)
			if err != nil {
				ctx.AbortWithError(410, err)
				return false
			}
			if len(list) == 0 { // avoid CLOSE_WAIT
				return true
			}
			for _, obj := range list {
				data, _ := json.Marshal(obj)
				ctx.Writer.Write(data)
			}
			ctx.Writer.Flush()
			resourceVersion = curVersion
		}
	})
}

func (res *ResourceHandler) AddFunc(obj any) {
	event := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: obj.(runtime.Object)}}
	res.fifo.Push(&event)
}

func (res *ResourceHandler) UpdateFunc(oldObj, newObj any) {
	event := metav1.WatchEvent{Type: "MODIFIED", Object: runtime.RawExtension{Object: newObj.(runtime.Object)}}
	res.fifo.Push(&event)
}

func (res *ResourceHandler) DeleteFunc(obj any) {
	event := metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: obj.(runtime.Object)}}
	res.fifo.Push(&event)
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
	return &ResourceHandler{GVR: gvr, fifo: NewResourceFifo()}
}
