package main

import "github.com/anhk/kube-relay/pkg/log"

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (app *App) Run(option *Option) error {
	for _, resource := range option.ResourceNames {
		log.Info("resource: %v", resource)
	}
	return nil
}
