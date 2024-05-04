package reader

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Registry interface {
	// Pull loads an image from a remote OCI registry
	Pull(ref string) (v1.Image, error)
}
