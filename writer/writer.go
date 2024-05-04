package writer

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Writer interface {
	Write(image v1.Image, imageRef, output, privateKeyPath string) error
}
