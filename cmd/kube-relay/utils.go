package main

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func processResource(arg string) *schema.GroupVersionResource {
	gvr := &schema.GroupVersionResource{}
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
