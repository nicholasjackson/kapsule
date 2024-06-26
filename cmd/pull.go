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

			if debug {
				logger.SetLevel(log.DebugLevel)
			}

			tag := args[0]

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

			logger.Info("Pulling image", "tag", tag, "output", outputFolder)
			r := reader.NewOCIRegistry(logger, registryUsername, registryPassword, insecure)
			i, err := r.Pull(tag)
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
				w := writer.NewOllamaWriter(logger, kp, outputFolder)
				err := w.Write(i, tag, decrypt, unzip)
				if err != nil {
					log.Error("Failed to write image to ollama", "path", outputFolder, "error", err)
					return
				}
			case "oci":
				if outputFolder == "" {
					log.Error("Output folder '--output-folder' must be specified for Ollama format")
					return
				}

				w := writer.NewPathWriter(logger, kp, outputFolder)
				err := w.Write(i, outputFolder, decrypt, unzip)
				if err != nil {
					log.Error("Failed to write image to path", "path", outputFolder, "error", err)
					return
				}
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
	pullCmd.Flags().BoolVarP(&unzip, "unzip", "", true, "Uncompresses layers when writing to disk")
	pullCmd.Flags().StringVarP(&registryUsername, "username", "", "", "Specify the username for the remote registry")
	pullCmd.Flags().StringVarP(&registryPassword, "password", "", "", "Specify the password for the remote registry")
	pullCmd.Flags().StringVarP(&encryptionKey, "encryption-key", "", "", "The encryption key to use for encrypting the image")
	pullCmd.Flags().StringVarP(&decryptionKey, "decryption-key", "", "", "The decryption key to use for encrypting the image, RSA private key")
	pullCmd.Flags().StringVarP(&encryptionVaultPath, "encryption-vault-path", "", "", "The path for the transit secrets engine in vault to use for encrypting and decrypting the image")
	pullCmd.Flags().StringVarP(&encryptionVaultKey, "encryption-vault-key", "", "", "The name of the key in vault to use for encrypting and decrypting the image")
	pullCmd.Flags().StringVarP(&encryptionVaultAuthToken, "encryption-vault-auth-token", "", "", "The vault token to use for accessing the encryption and decryption key")
	pullCmd.Flags().StringVarP(&encryptionVaultAuthAddr, "encryption-vault-addr", "", "", "The address of the vault server to use for accessing the encryption / decryption key")
	pullCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable logging in debug mode")

	return pullCmd
}
