package writer

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
type WriterImpl struct{}

func WriteToPath(image v1.Image, output string) error {
	var err error
	var p layout.Path

	p, err = layout.FromPath(output)
	if err != nil {
		// no index exists at the path, create a new index
		p, err = layout.Write(output, empty.Index)
		if err != nil {
			return err
		}
	}

	err = p.AppendImage(image)
	if err != nil {
		return err
	}

	return nil
}

func PushToRegistry(imageRef string, image v1.Image, username, password string) error {
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

	return remote.Write(ref, image, remote.WithAuth(auth))
}
