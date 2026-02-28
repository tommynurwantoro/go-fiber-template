package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "app",
}

func init() {
	rootCmd.AddCommand(RunService())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
