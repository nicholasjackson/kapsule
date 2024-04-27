package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var imageFile string
var output string

func newBuildCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build an OCI image from a model",
		Long:  `Build an OCI image for a model`,
		Args:  cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(2)),
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(os.Stdout)
			logger.SetReportTimestamp(false)

			ctx := args[0]
			out := args[1]

			log.Info("Building image", "modelfile", imageFile, "context", ctx, "output", out)
		},
	}

	buildCmd.Flags().StringVarP(&imageFile, "file", "f", "ModelFile", "Specify the model file for the build")
	buildCmd.Flags().StringVarP(&output, "output", "o", ".", "Specify the output directory for the built image")

	return buildCmd
}
