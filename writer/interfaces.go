package writer

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Writer interface {
	Write(image v1.Image, imageRef string, decrypt, unzip bool) error
	WriteEncrypted(image v1.Image, imageRef string) error
}
