package main

import "github.com/spf13/cobra"

var option = &Option{}

var rootCmd = &cobra.Command{
	Use: "kube-relay, to reduce the pressure of kube-apiserver when lot's of apps subscribing it",
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewApp().Run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&option.KubeConfig, "kubeconfig", "", "kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&option.ApiServer, "apiserver", "", "the address of apiserver")
}
