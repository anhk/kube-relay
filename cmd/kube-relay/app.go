package main

import (
	"github.com/anhk/kube-relay/pkg/log"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (app *App) Run(option *Option) error {
	for _, resource := range option.ResourceNames {
		gvr := processResource(resource)
		log.Info("resource=%v, group=%v, version=%v", gvr.Resource, gvr.Group, gvr.Version)
	}
	// Step 1# 检查所有资源都存在

	// Step 2# 启动Informer，等待Cache完成

	// Step 3# 启动侦听

	return nil
}
