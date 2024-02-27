package main

import (
	"github.com/anhk/kube-relay/pkg/log"
	"github.com/spf13/cobra"
)

var option = Option{}

var rootCmd = &cobra.Command{
	Use:     "kube-relay, to reduce the pressure of kube-apiserver when lot's of apps subscribing it",
	Example: "  kube-relay --resources services,endpointslice.discovery.k8s.io/v1",
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewApp().Run(&option)
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&option.KubeConfig, "kubeconfig", "", "kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&option.ApiServer, "apiserver", "", "the address of apiserver")

	rootCmd.PersistentFlags().StringArrayVar(&option.ResourceNames, "resources",
		[]string{"services", "endpointslices.discovery.k8s.io/v1"}, "resources to relay")

	rootCmd.PersistentFlags().IntVar(&log.Level, "loglevel", log.LEVEL_INFO, "log level")
	rootCmd.Execute()
}
