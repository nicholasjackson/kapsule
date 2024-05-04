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
	filePath    string
}

func NewPathWriter(logger *log.Logger, keyProvider crypto.KeyProvider, path string) *PathWriter {
	return &PathWriter{
		logger:      logger,
		keyProvider: keyProvider,
		filePath:    path,
	}
}

// WriteToPath writes the image to a local OCI image registry defined by output
func (pw *PathWriter) Write(image v1.Image, imageRef string, decypt, unzip bool) error {
	pw.logger.Info("Attempting to opening existing local path", "path", pw.filePath)
	p, err := pw.createOrOpenPath()
	if err != nil {
		return err
	}

	if decypt {
		pw.logger.Info("Decrypting layers with private key")

		pk, err := pw.keyProvider.PrivateKey()
		if err != nil {
			return fmt.Errorf("unable to get private key: %s", err)
		}

		// wrap the layers in a decrypted layer
		ei, err := wrapLayersWithDecryptedLayer(image, pk)
		if err != nil {
			return fmt.Errorf("unable to encrypt image: %s", err)
		}

		// replate the image with the encrypted image
		image = ei
	}

	// save the image
	err = p.AppendImage(image)
	if err != nil {
		return fmt.Errorf("unable to save image: %s", err)
	}

	if unzip {
		pw.logger.Info("Unzipping layers")
		err = unzipLayers(p, image)
		if err != nil {
			return fmt.Errorf("unable to unzip layers: %s", err)
		}
	}

	return nil
}

func (pw *PathWriter) WriteEncrypted(image v1.Image, imageRef string) error {
	pw.logger.Info("Attempting to opening existing local path", "path", pw.filePath)
	p, err := pw.createOrOpenPath()
	if err != nil {
		return err
	}

	pw.logger.Info("Encrypting layers with public key")
	pk, err := pw.keyProvider.PublicKey()
	if err != nil {
		return fmt.Errorf("unable to get public key: %s", err)
	}

	ei, err := wrapLayersWithEncryptedLayer(image, pk)
	if err != nil {
		return fmt.Errorf("unable to encrypt image: %s", err)
	}

	// replace the image with the encrypted image
	image = ei

	// we must save the image befoe we can update the annotations
	// the annotations contain the encrypted key that is used
	// to decrypt the image
	err = p.AppendImage(image)
	if err != nil {
		return err
	}

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

	return nil
}

func (pw *PathWriter) createOrOpenPath() (layout.Path, error) {
	p, err := layout.FromPath(pw.filePath)
	if err != nil {
		// no index exists at the path, create a new index
		pw.logger.Info("Path does not exist, creating new path", "path", pw.filePath)
		p, err = layout.Write(pw.filePath, empty.Index)
		if err != nil {
			return layout.Path(""), fmt.Errorf("unable to create new path: %s", err)
		}
	}

	return p, nil
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
