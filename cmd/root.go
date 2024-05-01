package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newPullCmd())
}

var rootCmd = &cobra.Command{
	Use:   "kapsule",
	Short: "Kapsule enables large language models to be bundled into OCI images",
	Long:  "Kapsule enables large language models to be bundled into OCI images",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
