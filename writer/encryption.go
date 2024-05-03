package writer

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/nicholasjackson/kapsule/crypto"
)

const (
	ENCRYPTION_KEY_ANNOTATION = "org.opencontainers.image.enc.keys.jwe"
	ENCRYPTION_KEY_OPTIONS    = "org.opencontainers.image.enc.pubopts"
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

// after writing an encrypted layer the encryption details used to encrypt the layer
// are stored in the annotations, this function reads the annotations and updates the
// image mnaifest with the encryption details
func appendEncyptedLayerAnnotations(image v1.Image) (v1.Image, error) {
	new := empty.Image

	// get the layers from the image
	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("unable to get layers from image: %s", err)
	}

	// iterate over the layers and update the annotations
	for _, l := range layers {
		// get the annotations
		ann, err := l.(*crypto.EncryptedLayer).Annotations()
		if err != nil {
			return nil, fmt.Errorf("unable to get annotations from encrypted layer: %s", err)
		}

		new, err = mutate.Append(new, mutate.Addendum{Layer: l, Annotations: ann})
		if err != nil {
			return nil, fmt.Errorf("unable to append layer to image: %s", err)
		}
	}

	return new, nil
}

// wrapLayersWithDecryptedLayer wraps each layer in the image with a decrypted layer
func wrapLayersWithDecryptedLayer(image v1.Image, privateKeyFile string) (v1.Image, error) {
	new := empty.Image

	key, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key file: %s", err)
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("unable to get layers from image: %s", err)
	}

	for _, l := range layers {
		// check the media type
		mt, err := l.MediaType()
		if err != nil {
			return nil, fmt.Errorf("unable to get media type from layer: %s", err)
		}

		// if the layer is encrypted we need to decrypt it
		if strings.HasSuffix(string(mt), "+enc") {
			// get the annotation, this is needed to decrypt the image
			ann := getLayerAnnotationsFromImage(image, l)
			if ann[ENCRYPTION_KEY_ANNOTATION] == "" || ann[ENCRYPTION_KEY_OPTIONS] == "" {
				return nil, fmt.Errorf("layer is encrypted but missing encryption annotations")
			}

			fmt.Println("encrypted layer")
			dl, err := crypto.NewDecryptedLayer(l, key, ann)
			if err != nil {
				return nil, fmt.Errorf("unable to create decrypted layer: %s", err)
			}

			// replace the layer
			l = dl

		}

		// add the layer back to the image
		new, err = mutate.AppendLayers(new, l)
		if err != nil {
			return nil, fmt.Errorf("unable to append layer to image: %s", err)
		}
	}

	return new, nil
}

func getLayerAnnotationsFromImage(image v1.Image, l v1.Layer) map[string]string {
	mf, err := image.Manifest()
	if err != nil {
		return map[string]string{}
	}

	for _, ml := range mf.Layers {
		d, err := l.Digest()
		if err != nil {
			continue
		}

		if ml.Digest == d {
			return ml.Annotations
		}
	}

	return map[string]string{}
}
