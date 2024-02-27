package main

type Option struct {
	KubeConfig string
	ApiServer  string

	ResourceNames []string
	Port          uint16 // Listen Port
}
