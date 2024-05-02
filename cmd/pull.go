package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/nicholasjackson/kapsule/reader"
	"github.com/nicholasjackson/kapsule/writer"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull an OCI image from a remote registry",
		Long:  ``,
		Args:  cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(os.Stdout)
			logger.SetReportTimestamp(false)

			tag := args[0]

			logger.Info("Pulling image", "tag", tag, "output", outputFolder)

			r := reader.ReaderImpl{}
			i, err := r.PullFromRegistry(tag, registryUsername, registryPassword)
			if err != nil {
				log.Error("Failed to pull image", "error", err)
				return
			}

			// write the image to the output
			switch outputFormat {
			case "ollama":
				if outputFolder == "" {
					log.Error("Output folder '--output-folder' must be specified for Ollama format")
					return
				}

				err := writer.WriteToOllama(i, tag, outputFolder)
				if err != nil {
					log.Error("Failed to write image to ollama", "path", outputFolder, "error", err)
					return
				}
			case "oci":
				if outputFolder == "" {
					err := writer.WriteToPath(i, outputFolder, encryptionKey)
					if err != nil {
						log.Error("Failed to write image to path", "path", outputFolder, "error", err)
						return
					}
				}

				log.Error("You must specify --output when pulling images")
				return
			default:
				log.Error("Unsupported format", "format", outputFormat)
				return
			}
		},
	}

	pullCmd.Flags().StringVarP(&outputFormat, "format", "", "oci", "Specify the output format for the built image, defaults to OCI image format, options: [ollama, oci]")
	pullCmd.Flags().StringVarP(&outputFolder, "output", "o", "", "Specify the output folder for the built image, if not specified the image will be pushed to a remote registry")
	pullCmd.Flags().BoolVarP(&insecure, "insecure", "", false, "Push to an insecure registry")
	pullCmd.Flags().StringVarP(&registryUsername, "username", "", "", "Specify the username for the remote registry")
	pullCmd.Flags().StringVarP(&registryPassword, "password", "", "", "Specify the password for the remote registry")
	pullCmd.Flags().StringVarP(&encryptionKey, "encryption-key", "", "", "The encryption key to use for encrypting the image")
	pullCmd.Flags().StringVarP(&decryptionKey, "decryption-key", "", "", "The decryption key to use for encrypting the image, RSA private key")
	pullCmd.Flags().StringVarP(&encryptionVaultKey, "encryption-vault-key", "", "", "The path to the exportable encryption key in vault to use for encrypting the image")
	pullCmd.Flags().StringVarP(&encryptionVaultAuthToken, "encryption-vault-auth-token", "", "", "The vault token to use for accessing the encryption key")
	pullCmd.Flags().StringVarP(&encryptionVaultAuthAddr, "encryption-vault-addr", "", "", "The address of the vault server to use for accessing the encryption key")

	return pullCmd
}
