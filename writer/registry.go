package writer

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type OCIRegistry struct {
	logger *log.Logger
}

func NewOCIRegistry(logger *log.Logger) *OCIRegistry {
	return &OCIRegistry{
		logger: logger,
	}
}

// Push pushes the given image to a remote OCI image registry
func (r *OCIRegistry) Push(imageRef string, image v1.Image, username, password, publicKeyPath string) error {
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
		r.logger.Info("Encrypting layers with public key", "publicKeyPath", publicKeyPath)

		ei, err := wrapLayersWithEncryptedLayer(image, publicKeyPath)
		if err != nil {
			return fmt.Errorf("unable to encrypt image: %s", err)
		}

		// replate the image with the encrypted image
		image = ei
	}

	// remote.WithProgress to write the image with progress
	r.logger.Info("Pushing image", "imageRef", imageRef)
	err = remote.Write(ref, image, remote.WithAuth(auth), remote.WithProgress(r.progressReport()))
	if err != nil {
		return fmt.Errorf("unable to write image to registry: %s", err)
	}

	// if we are encrypting the image we need to update the annotations
	// as they contain information that is needed to decrypt the image
	if publicKeyPath != "" {
		r.logger.Info("Updating layers with encryption details")

		newImage, err := appendEncyptedLayerAnnotations(image)
		if err != nil {
			return fmt.Errorf("unable to update annotations: %s", err)
		}

		r.logger.Info("Updating remote image", "imageRef", imageRef)
		err = remote.Write(ref, newImage, remote.WithAuth(auth), remote.WithProgress(r.progressReport()))
		if err != nil {
			return fmt.Errorf("unable to write image to registry: %s", err)
		}
	}

	return nil
}

func (r *OCIRegistry) progressReport() chan v1.Update {
	ch := make(chan v1.Update, 1)
	total := int64(0)
	completed := int64(0)

	t := time.AfterFunc(5*time.Second, func() {
		percentage := "0%"

		if completed > 0 && total > 0 {
			p := int((completed / total) * 100)

			percentage = fmt.Sprintf("%d%%", p)

			r.logger.Info("Pushing image", "complete", percentage, "total", total, "completed", completed)
		}
	})

	go func() {
		for {
			update, ok := <-ch
			if !ok {
				r.logger.Info("Pushing image complete")
				t.Stop()
				return
			}

			total = update.Total
			completed = update.Complete
		}
	}()

	return ch
}
