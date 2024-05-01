package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type DecryptedLayer struct {
	layer v1.Layer
	key   []byte
}

func (dl *DecryptedLayer) Compressed() (io.ReadCloser, error) {
	r, err := dl.layer.Compressed()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dl.key)
	if err != nil {
		return nil, err
	}

	var iv [aes.BlockSize]byte
	if _, err := io.ReadFull(rand.Reader, iv[:]); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBDecrypter(block, iv[:])

	reader := &cipher.StreamReader{S: stream, R: r}

	return io.NopCloser(reader), nil
}
