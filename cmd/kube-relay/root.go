package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "kube-relay, to reduce the pressure of kube-apiserver when lot's of apps subscribing it",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
