package reader

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Reader interface {
	// ReadFromPath loads an image from a local OCI registry
	ReadFromPath(path string) (v1.Image, error)
	// PullFromRegistry loads an image from a remote OCI registry
	PullFromRegistry(ref, username, password string) (v1.Image, error)
}

type ReaderImpl struct{}

// PullFromRegistry loads an image from a remote OCI registry
func (r *ReaderImpl) PullFromRegistry(imageRef, username, password string) (v1.Image, error) {
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
		return nil, err
	}

	auth := authn.FromConfig(*cfg)

	return remote.Image(ref, remote.WithAuth(auth))
}
