package main

import (
	"sync"

	"github.com/anhk/kube-relay/pkg/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type App struct {
	kubeClient dynamic.Interface

	resMap sync.Map
}

func NewApp() *App {
	return &App{}
}

func (app *App) Run(option *Option) (err error) {
	if app.kubeClient, err = CreateDynamicClient(option.KubeConfig, option.ApiServer); err != nil {
		return err
	}
	var listCached []cache.InformerSynced
	for _, resName := range option.ResourceNames {
		var resHandler = NewResourceHandler(processResource(resName))
		app.resMap.Store(resHandler.GVR, resHandler)
		listCached = append(listCached, resHandler.RunWithDynamicClient(app.kubeClient))
	}

	if ok := cache.WaitForCacheSync(wait.NeverStop, listCached...); ok {
		log.Info("cache ok")
	}

	select {}
}
