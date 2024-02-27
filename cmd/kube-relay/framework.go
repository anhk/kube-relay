package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Services
// - /api/v1/services
// - /api/v1/namespaces/default/services?limit=500
// - /api/v1/namespaces/default/services/echoserver

// Deployments
// - /apis/apps/v1/deployments
// - /apis/apps/v1/namespaces/default/deployments
// - /apis/apps/v1/namespaces/default/deployments/my-deploy

// EndpointSlices
// - /apis/discovery.k8s.io/v1/endpointslices?limit=500
// - /apis/discovery.k8s.io/v1/namespaces/default/endpointslices
// - /apis/discovery.k8s.io/v1/namespaces/default/endpointslices/kubernetes

// Watch
// - url?watch=true&resourceVersion=10245
// - url?watch=1&resourceVersion=10245

// Example
// - /api/v1/namespaces/kube-system/configmaps?
// 	   allowWatchBookmarks=true
//     &fieldSelector=metadata.name%3Dextension-apierver-authentication
//     &resourceVersion=773059
//     &timeout=9m9s
//     &timeoutSeconds=549
//     &watch=true

type Router struct {
	r *gin.Engine
}

func NewRouter() *Router {
	r := &Router{r: gin.Default()}
	r.r.GET("/api", r.apiList)
	r.r.GET("/api/v1", r.apiCoreList)
	r.r.GET("/apis", r.apisList)
	return r
}

// 返回核心API列表
func (r *Router) apiList(ctx *gin.Context) {
	ctx.Writer.Write([]byte(`{
	"kind": "APIVersions",
	"versions": [
	  "v1"
	],
	"serverAddressByClientCIDRs": [
	  {
		"clientCIDR": "0.0.0.0/0",
		"serverAddress": "192.168.1.222:6443"
	  }
	]
  }`))
}

func (r *Router) apiCoreList(ctx *gin.Context) {
	ctx.Writer.Write([]byte(`{
		"kind": "APIResourceList",
		"groupVersion": "v1",
		"resources": [
			{
				"name": "services",
				"singularName": "service",
				"namespaced": true,
				"kind": "Service",
				"verbs": [
				  "get",
				  "list",
				  "watch"
				],
				"storageVersionHash": "0/CO1lhkEBI="
			  },
		]
	}`))
}

// 返回非核心API列表
func (r *Router) apisList(ctx *gin.Context) {
	ctx.Writer.Write([]byte(`{
		"kind": "APIGroupList",
		"apiVersion": "v1",
		"groups": [
		  {
			"name": "discovery.k8s.io",
			"versions": [
			  {
				"groupVersion": "discovery.k8s.io/v1",
				"version": "v1"
			  }
			],
			"preferredVersion": {
			  "groupVersion": "discovery.k8s.io/v1",
			  "version": "v1"
			}
		  }
		]
	  }`))
}

func (r *Router) Watch(gvr *schema.GroupVersionResource, fn gin.HandlerFunc) {
	if gvr.Group == "" {
		r.r.GET(fmt.Sprintf("/api/%v/%v", gvr.Version, gvr.Resource), fn)
		r.r.GET(fmt.Sprintf("/api/%v/namespaces/:namespace/%v", gvr.Version, gvr.Resource), fn)
		r.r.GET(fmt.Sprintf("/api/%v/namespaces/:namespace/%v/:name", gvr.Version, gvr.Resource), fn)
	} else {
		r.r.GET(fmt.Sprintf("/apis/%v/%v/%v", gvr.Group, gvr.Version, gvr.Resource), fn)
		r.r.GET(fmt.Sprintf("/apis/%v/%v/namespaces/:namespace/%v", gvr.Group, gvr.Version, gvr.Resource), fn)
		r.r.GET(fmt.Sprintf("/apis/%v/%v/namespaces/:namespace/%v/:name", gvr.Group, gvr.Version, gvr.Resource), fn)
	}
}

func (r *Router) Run(port uint16) error {
	return r.r.Run(fmt.Sprintf(":%v", port))
}
