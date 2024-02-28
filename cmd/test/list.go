package main

import (
	"context"

	"github.com/anhk/kube-relay/pkg/k8s"
	"github.com/anhk/kube-relay/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func listService(option *Option) {
	kubeClient, err := k8s.CreateKubeClient(option.KubeConfig, option.ApiServer)
	PanicIf(err)

	svcList, err := kubeClient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	PanicIf(err)

	for _, svc := range svcList.Items {
		log.Info("svc: %+v", svc)
	}
}
