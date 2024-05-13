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
var encryptionVaultPath string
var encryptionVaultKey string
var encryptionVaultAuthToken string
var encryptionVaultAuthAddr string
var encryptionVaultAuthNamespace string
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

			// are we using encryption, if so build the key provider
			kp, err := getKeyProvider(
				logger,
				encryptionKey,
				decryptionKey,
				encryptionVaultKey,
				encryptionVaultPath,
				encryptionVaultAuthToken,
				encryptionVaultAuthAddr,
				encryptionVaultAuthNamespace)

			if err != nil {
				log.Error("Failed to create key provider", "error", err)
				return
			}

			decrypt := false
			if decryptionKey != "" || encryptionVaultKey != "" {
				decrypt = true
			}

			encrypt := false
			if encryptionKey != "" || encryptionVaultKey != "" {
				encrypt = true
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

				w := writer.NewOllamaWriter(logger, kp, outputFolder)
				err := w.Write(i, tag, decrypt, unzip)
				if err != nil {
					log.Error("Failed to write image to ollama", "path", outputFolder, "error", err)
					return
				}
			case "oci":
				if outputFolder != "" {
					w := writer.NewPathWriter(logger, kp, outputFolder)

					var err error
					if encrypt {
						err = w.WriteEncrypted(i, tag)
					} else {
						err = w.Write(i, tag, decrypt, unzip)
					}

					if err != nil {
						log.Error("Failed to write image to path", "path", outputFolder, "error", err)
						return
					}
				} else {
					w := writer.NewOCIRegistry(logger, kp, registryUsername, registryPassword, insecure)

					var err error
					if encrypt {
						err = w.WriteEncrypted(i, tag)
					} else {
						err = w.Write(i, tag, decrypt, unzip)
					}

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
	buildCmd.Flags().StringVarP(&encryptionVaultPath, "encryption-vault-path", "", "", "The path to the transit secrets endpoint for encrypting and decryupting the image")
	buildCmd.Flags().StringVarP(&encryptionVaultKey, "encryption-vault-key", "", "", "The name of exportable encryption key in Vault to use for encrypting and decrypting the image")
	buildCmd.Flags().StringVarP(&encryptionVaultAuthToken, "encryption-vault-auth-token", "", "", "The vault token to use for accessing the encryption key")
	buildCmd.Flags().StringVarP(&encryptionVaultAuthAddr, "encryption-vault-addr", "", "", "The address of the vault server to use for accessing the encryption key")
	buildCmd.Flags().StringVarP(&encryptionVaultAuthNamespace, "encryption-vault-namespace", "", "", "The namespace for the vault server to use for accessing the encryption key")
	buildCmd.Flags().BoolVarP(&unzip, "unzip", "", true, "Uncompresses layers when writing to disk")
	buildCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable logging in debug mode")

	return buildCmd
}
