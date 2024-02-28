package main

import (
	"github.com/anhk/kube-relay/pkg/log"
	"github.com/spf13/cobra"
)

type Option struct {
	KubeConfig string
	ApiServer  string
}

var option = Option{}

var rootCmd = &cobra.Command{
	Use:          "test",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("==================")
		listService(&option)
		log.Info("==================")
		watchService(&option)
	},
}

func main() {
	rootCmd.PersistentFlags().StringVar(&option.KubeConfig, "kubeconfig", "", "kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&option.ApiServer, "apiserver", "", "the address of apiserver")
	rootCmd.Execute()
}

func PanicIf(e any) {
	if e != nil {
		panic(e)
	}
}
