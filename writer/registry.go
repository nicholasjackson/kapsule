package writer

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// PushToRegistry pushes the given image to a remote OCI image registry
func PushToRegistry(imageRef string, image v1.Image, username, password, publicKeyPath string) error {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		panic(err)
	}

	b := authn.Basic{
		Username: username,
		Password: password,
	}

	cfg, err := b.Authorization()
	if err != nil {
		return err
	}

	auth := authn.FromConfig(*cfg)

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

	// remote.WithProgress to write the image with progress
	err = remote.Write(ref, image, remote.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("unable to write image to registry: %s", err)
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
		err = remote.Write(ref, newImage, remote.WithAuth(auth))
		if err != nil {
			return fmt.Errorf("unable to write image to registry: %s", err)
		}
	}

	return nil
}
