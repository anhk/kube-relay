package main

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func processResource(arg string) schema.GroupVersionResource {
	gvr := schema.GroupVersionResource{}
	if i := strings.Index(arg, "/"); i >= 0 {
		gvr.Version = arg[i+1:]
		arg = arg[:i]
	} else {
		gvr.Version = "v1"
	}

	if i := strings.Index(arg, "."); i >= 0 {
		gvr.Resource = arg[:i]
		gvr.Group = arg[i+1:]
	} else {
		gvr.Resource = arg
	}
	return gvr
}

func CreateDynamicClient(kubeConfigFile, apiServer string) (dynamic.Interface, error) {
	var clientconfig *rest.Config
	var err error

	if len(apiServer) != 0 || len(kubeConfigFile) != 0 {
		if clientconfig, err = clientcmd.BuildConfigFromFlags(apiServer, kubeConfigFile); err != nil {
			return nil, fmt.Errorf("Failed to build configuration from CLI: " + err.Error())
		}
	} else if clientconfig, err = rest.InClusterConfig(); err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: " + err.Error())
	}
	clientconfig.QPS = 1000
	clientconfig.Burst = 5000
	return dynamic.NewForConfig(clientconfig)
}

// objectToUnstructured: 对象转换为 *unstructured.Unstructured
func objectToUnstructured(obj any) *unstructured.Unstructured {
	utd, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return &unstructured.Unstructured{Object: utd}
}

// // unstructuredToObject: 将 *unstructured.Unstructured 转换为对象
// func unstructuredToObject[T any](utd *unstructured.Unstructured, obj *T) *T {
// 	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(utd.UnstructuredContent(), obj)
// 	return obj
// }
