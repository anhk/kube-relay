package main

import (
	"fmt"

	"github.com/anhk/kube-relay/pkg/log"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type App struct {
	kubeClient    *kubernetes.Clientset
	dynamicClient dynamic.Interface
	resMap        map[schema.GroupVersionResource]*ResourceHandler

	Engine *gin.Engine
}

func NewApp() *App {
	return &App{resMap: make(map[schema.GroupVersionResource]*ResourceHandler)}
}

func (app *App) Run(option *Option) (err error) {
	// Step. 1# 创建Kubernetes客户端
	if app.kubeClient, err = CreateKubeClient(option.KubeConfig, option.ApiServer); err != nil {
		return err
	}

	// Step. 2# 预处理资源
	for _, resName := range option.ResourceNames {
		var resHandler = NewResourceHandler(processResource(resName))
		if err := resHandler.GetInfoByKubeClient(app.kubeClient); err != nil {
			return err
		}
		app.resMap[resHandler.GVR] = resHandler
	}

	// Step. 2# 建立动态客户端
	if app.dynamicClient, err = CreateDynamicClient(option.KubeConfig, option.ApiServer); err != nil {
		return err
	}

	// Step. 3# 同步数据直到完成缓存
	var listCached []cache.InformerSynced
	for _, resHandler := range app.resMap {
		listCached = append(listCached, resHandler.RunWithDynamicClient(app.dynamicClient))
	}

	if ok := cache.WaitForCacheSync(wait.NeverStop, listCached...); ok {
		log.Info("cache ok")
	}

	// Step. 4# 启动HTTP(s)侦听
	gin.SetMode(gin.ReleaseMode)
	app.Engine = gin.Default()

	for gvr, resHandler := range app.resMap {
		app.SetWatchFunc(&gvr, resHandler.WatchFunc)
	}
	app.SetApiListFunc()
	return app.Engine.Run(fmt.Sprintf(":%v", option.Port))
}

func (app *App) SetApiListFunc() {
	app.Engine.GET("/api", app.APIVersion)
	app.Engine.GET("/apis", app.APIGroupList)
	app.Engine.GET("/api/v1", app.APICoreResourceList)
}

// 设置Watch资源的回调函数
func (app *App) SetWatchFunc(gvr *schema.GroupVersionResource, fn gin.HandlerFunc) {
	if gvr.Group == "" {
		app.Engine.GET(fmt.Sprintf("/api/%v/%v", gvr.Version, gvr.Resource), fn)
		app.Engine.GET(fmt.Sprintf("/api/%v/namespaces/:namespace/%v", gvr.Version, gvr.Resource), fn)
		app.Engine.GET(fmt.Sprintf("/api/%v/namespaces/:namespace/%v/:name", gvr.Version, gvr.Resource), fn)
	} else {
		app.Engine.GET(fmt.Sprintf("/apis/%v/%v/%v", gvr.Group, gvr.Version, gvr.Resource), fn)
		app.Engine.GET(fmt.Sprintf("/apis/%v/%v/namespaces/:namespace/%v", gvr.Group, gvr.Version, gvr.Resource), fn)
		app.Engine.GET(fmt.Sprintf("/apis/%v/%v/namespaces/:namespace/%v/:name", gvr.Group, gvr.Version, gvr.Resource), fn)
	}
}

// 核心API版本
func (app *App) APIVersion(ctx *gin.Context) {
	ctx.Writer.Write([]byte(`{"kind": "APIVersions", "versions": [ "v1" ]}`))
}

// 核心API列表
func (app *App) APICoreResourceList(ctx *gin.Context) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = "v1"
	for gvr := range app.resMap {
		if gvr.Group != "" { // 非核心API
			continue
		}
		res := metav1.APIResource{
			Name:               gvr.Resource,
			SingularName:       "service",                        // TODO
			Namespaced:         true,                             // TODO
			ShortNames:         []string{"svc"},                  // TODO
			Kind:               "Service",                        // TODO
			Verbs:              []string{"get", "list", "watch"}, // TODO
			StorageVersionHash: "",                               // TODO
		}
		apiResourceList.APIResources = append(apiResourceList.APIResources, res)
	}

	ctx.JSON(200, apiResourceList)
}

// 返回非核心API列表
func (app *App) APIGroupList(ctx *gin.Context) {
	apiGroupList := &metav1.APIGroupList{}
	ctx.JSON(200, apiGroupList)
}
