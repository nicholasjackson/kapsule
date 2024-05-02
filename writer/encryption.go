package writer

import (
	"fmt"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/nicholasjackson/kapsule/crypto"
)

// wrapLayersWithEncryptedLayer wraps each layer in the image with an encrypted layer
// and returns the new image
func wrapLayersWithEncryptedLayer(i v1.Image, publicKeyFile string) (v1.Image, error) {
	base := empty.Image

	layers, err := i.Layers()
	if err != nil {
		return nil, fmt.Errorf("unable to get layers from image: %s", err)
	}

	pubKeyBytes, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key file: %s", err)
	}

	for _, l := range layers {
		el, err := crypto.NewEncryptedLayer(l, pubKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("unable to create encrypted layer: %s", err)
		}

		base, err = mutate.AppendLayers(base, el)
		if err != nil {
			return nil, fmt.Errorf("unable to append encrypted layer: %s", err)
		}
	}

	return base, nil
}
