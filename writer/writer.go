package writer

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

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

func WriteToRegistry(imageRef string, image v1.Image, username, password string) error {
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
