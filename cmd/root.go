package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newBuildCmd())
}

var rootCmd = &cobra.Command{
	Use:   "llm-image",
	Short: "LLM Image enables large language models to be bundled into OCI images",
	Long:  "LLM Image enables large language models to be bundled into OCI images",
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
