package writer

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/log"
	"github.com/nicholasjackson/kapsule/crypto"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
)

// WriterImpl is a concrete implementation of the Writer interface
type PathWriter struct {
	logger      *log.Logger
	keyProvider crypto.KeyProvider
}

func NewPathWriter(logger *log.Logger, keyProvider crypto.KeyProvider) *PathWriter {
	return &PathWriter{
		logger:      logger,
		keyProvider: keyProvider,
	}
}

// WriteToPath writes the image to a local OCI image registry defined by output
func (pw *PathWriter) Write(image v1.Image, output, string, decypt, unzip bool) error {
	var err error
	var p layout.Path

	if unzip && publicKeyPath != "" {
		return fmt.Errorf("unzip is not supported when encrypting the image")
	}

	pw.logger.Info("Attempting to opening existing local path", "path", output)
	p, err = layout.FromPath(output)
	if err != nil {
		// no index exists at the path, create a new index
		pw.logger.Info("Path does not exist, creating new path", "path", output)
		p, err = layout.Write(output, empty.Index)
		if err != nil {
			return err
		}
	}

	// if we have a public key, we need to encrypt the image
	// we do this by wrapping the image in a layers with an
	// encrypted layer
	if publicKeyPath != "" {
		pw.logger.Info("Encrypting layers with public key", "publicKeyPath", publicKeyPath)

		ei, err := wrapLayersWithEncryptedLayer(image, publicKeyPath)
		if err != nil {
			return fmt.Errorf("unable to encrypt image: %s", err)
		}

		// replate the image with the encrypted image
		image = ei
	}

	if privateKeyPath != "" {
		pw.logger.Info("Decrypting layers with private key", "privateKeyPath", privateKeyPath)

		// wrap the layers in a decrypted layer
		ei, err := wrapLayersWithDecryptedLayer(image, privateKeyPath)
		if err != nil {
			return fmt.Errorf("unable to encrypt image: %s", err)
		}

		// replate the image with the encrypted image
		image = ei
	}

	err = p.AppendImage(image)
	if err != nil {
		return err
	}

	// if we are encrypting the image we need to update the annotations
	// as they contain information that is needed to decrypt the image
	if publicKeyPath != "" {
		pw.logger.Info("Adding annotations from encryption process to manifest")

		newImage, err := appendEncyptedLayerAnnotations(image)
		if err != nil {
			return fmt.Errorf("unable to update annotations: %s", err)
		}

		digest, err := image.Digest()
		if err != nil {
			return fmt.Errorf("unable to get digest: %s", err)
		}

		pw.logger.Info("Updating image")
		err = p.ReplaceImage(newImage, match.Digests(digest))
		if err != nil {
			return err
		}
	}

	// do we need up unzip the layers?
	if unzip {
		pw.logger.Info("Unzipping layer content")

		err = unzipLayers(p, image)
		if err != nil {
			return fmt.Errorf("unable to unzip layers: %s", err)
		}
	}

	return nil
}

// unzip the blob content of each layer
func unzipLayers(p layout.Path, image v1.Image) error {
	layers, err := image.Layers()
	if err != nil {
		return fmt.Errorf("unable to get layers from image: %s", err)
	}

	for _, l := range layers {
		d, err := l.Digest()
		if err != nil {
			return fmt.Errorf("unable to get digest: %s", err)
		}

		rc, err := p.Blob(d)
		if err != nil {
			return fmt.Errorf("unable to get blob: %s", err)
		}

		// create a temp file to write the uncompressed layer to
		tempFile, err := os.CreateTemp("", "uncompressed_layer_")
		if err != nil {
			return fmt.Errorf("unable to create temporary file: %s", err)
		}
		defer tempFile.Close()
		defer os.Remove(tempFile.Name())

		gzr, err := gzip.NewReader(rc)
		if err != nil {
			return fmt.Errorf("unable to create gzip reader: %s", err)
		}

		_, err = io.Copy(tempFile, gzr)
		if err != nil {
			return fmt.Errorf("unable to copy contents to temporary file: %s", err)
		}

		// close the gzip reader
		gzr.Close()

		// remove the old blob
		err = p.RemoveBlob(d)
		if err != nil {
			return fmt.Errorf("unable to remove compressed layer: %s", err)
		}

		tempFile.Seek(0, 0)

		// write the new blob
		err = p.WriteBlob(d, tempFile)
		if err != nil {
			return fmt.Errorf("unable to write uncompressed layer: %s", err)
		}

	}

	return nil
}
