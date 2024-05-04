package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/nicholasjackson/kapsule/builder"
	"github.com/nicholasjackson/kapsule/writer"
	"github.com/spf13/cobra"
)

var modelFile string
var tag string
var outputFormat string
var outputFolder string
var insecure bool
var registryUsername string
var registryPassword string
var encryptionKey string
var decryptionKey string
var encryptionVaultKey string
var encryptionVaultAuthToken string
var encryptionVaultAuthAddr string
var unzip bool
var debug bool

func newBuildCmd() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build an OCI image from a model",
		Long: `
			Builds an OCI image for a model using the specified context and output format.
			`,
		Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(os.Stdout)
			logger.SetReportTimestamp(false)

			if debug {
				logger.SetLevel(log.DebugLevel)
			}

			ctx := args[0]

			logger.Info("Building image", "modelfile", modelFile, "context", ctx, "output", outputFolder, "format", outputFormat, "tag", tag)

			b := builder.NewBuilder()
			i, err := b.Build(modelFile, ctx)
			if err != nil {
				log.Error("Failed to build image", "error", err)
				return
			}

			// write the image to the output
			switch outputFormat {
			case "ollama":
				if outputFolder == "" {
					log.Error("Output folder '--output-folder' must be specified for Ollama format")
					return
				}

				w := writer.NewOllamaWriter(logger)
				err := w.Write(i, tag, outputFolder, decryptionKey)
				if err != nil {
					log.Error("Failed to write image to ollama", "path", outputFolder, "error", err)
					return
				}
			case "oci":
				if outputFolder != "" {
					w := writer.NewPathWriter(logger)
					err := w.Write(i, outputFolder, encryptionKey, decryptionKey, unzip)
					if err != nil {
						log.Error("Failed to write image to path", "path", outputFolder, "error", err)
						return
					}
				} else {
					w := writer.NewOCIRegistry(logger)
					err := w.Push(tag, i, registryUsername, registryPassword, encryptionKey)
					if err != nil {
						log.Error("Failed to push image to remote registry", "error", err)
						return
					}
				}
			default:
				log.Error("Unsupported format", "format", outputFormat)
				return
			}
		},
	}

	buildCmd.Flags().StringVarP(&modelFile, "file", "f", "ModelFile", "Specify the model file for the build")
	buildCmd.Flags().StringVarP(&tag, "tag", "t", "", "Specify the tag for the built image i.e. docker.io/nicholasjackson/llm_test:latest")
	buildCmd.Flags().StringVarP(&outputFormat, "format", "", "oci", "Specify the output format for the built image, defaults to OCI image format, options: [ollama, oci]")
	buildCmd.Flags().StringVarP(&outputFolder, "output", "o", "", "Specify the output folder for the built image, if not specified the image will be pushed to a remote registry")
	buildCmd.Flags().BoolVarP(&insecure, "insecure", "", false, "Push to an insecure registry")
	buildCmd.Flags().StringVarP(&registryUsername, "username", "", "", "Specify the username for the remote registry")
	buildCmd.Flags().StringVarP(&registryPassword, "password", "", "", "Specify the password for the remote registry")
	buildCmd.Flags().StringVarP(&encryptionKey, "encryption-key", "", "", "The encryption key to use for encrypting the image, RSA public key")
	buildCmd.Flags().StringVarP(&decryptionKey, "decryption-key", "", "", "The decryption key to use for encrypting the image, RSA private key")
	buildCmd.Flags().StringVarP(&encryptionVaultKey, "encryption-vault-key", "", "", "The path to the exportable encryption key in vault to use for encrypting the image")
	buildCmd.Flags().StringVarP(&encryptionVaultAuthToken, "encryption-vault-auth-token", "", "", "The vault token to use for accessing the encryption key")
	buildCmd.Flags().StringVarP(&encryptionVaultAuthAddr, "encryption-vault-addr", "", "", "The address of the vault server to use for accessing the encryption key")
	buildCmd.Flags().BoolVarP(&unzip, "unzip", "", true, "Uncompresses layers when writing to disk")
	buildCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable logging in debug mode")

	return buildCmd
}
