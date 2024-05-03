package writer

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
)

type Writer interface {
	// Write to path writes the image to a local OCI image registry defined by output
	// if no existing regitry exists at the output path, WriteToPath scaffolds a new
	// regitstry before writing the image
	WriteToPath(image v1.Image, output string) error
	// PushToRegistry pushes the given image to a remote OCI image registry
	PushToRegistry(imageRef string, image v1.Image, username, password string) error
}

// WriterImpl is a concrete implementation of the Writer interface
type WriterImpl struct {
}

// WriteToPath writes the image to a local OCI image registry defined by output
func WriteToPath(image v1.Image, output, publicKeyPath, privateKeyPath string, unzip bool) error {
	var err error
	var p layout.Path

	if unzip && publicKeyPath != "" {
		return fmt.Errorf("unzip is not supported when encrypting the image")
	}

	p, err = layout.FromPath(output)
	if err != nil {
		// no index exists at the path, create a new index
		p, err = layout.Write(output, empty.Index)
		if err != nil {
			return err
		}
	}

	// if we have a public key, we need to encrypt the image
	// we do this by wrapping the image in a layers with an
	// encrypted layer
	if publicKeyPath != "" {
		fmt.Println("Encrypting image")
		ei, err := wrapLayersWithEncryptedLayer(image, publicKeyPath)
		if err != nil {
			return fmt.Errorf("unable to encrypt image: %s", err)
		}

		// replate the image with the encrypted image
		image = ei
	}

	if privateKeyPath != "" {
		// wrap the layers in a decrypted layer
		fmt.Println("Decrypting image")
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
		fmt.Println("Updating annotations")
		newImage, err := appendEncyptedLayerAnnotations(image)
		if err != nil {
			return fmt.Errorf("unable to update annotations: %s", err)
		}

		fmt.Println("Writing updated image")
		digest, err := image.Digest()
		if err != nil {
			return fmt.Errorf("unable to get digest: %s", err)
		}

		err = p.ReplaceImage(newImage, match.Digests(digest))
		if err != nil {
			return err
		}
	}

	// are we decrypting the image
	// do we need up unzip the layers?
	if unzip {
		fmt.Println("Unzipping layers")
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
