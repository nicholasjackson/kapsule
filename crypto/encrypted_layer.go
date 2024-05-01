package crypto

import (
	"fmt"
	"io"

	"crypto/sha256"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type EncryptedLayer struct {
	r   io.Reader
	key []byte
}

func NewEncryptedLayer(r io.Reader, key []byte) *EncryptedLayer {
	return &EncryptedLayer{
		r:   r,
		key: key,
	}
}

func (el *EncryptedLayer) Digest() (v1.Hash, error) {
	// You might want to cache this value, as calculating the digest of the encrypted data could be expensive
	h := sha256.New()
	if _, err := io.Copy(h, el.r); err != nil {
		return v1.Hash{}, err
	}

	return v1.NewHash(fmt.Sprintf("sha256:%x", h.Sum(nil)))
}

func (el *EncryptedLayer) DiffID() (v1.Hash, error) {
	// DiffID of the encrypted layer is not the same as the original layer
	return el.Digest()
}

func (el *EncryptedLayer) Compressed() (io.ReadCloser, error) {
	reader, err := EncryptedReader(el.key, el.r)
	if err != nil {
		return nil, fmt.Errorf("could not create encrypted reader: %w", err)
	}

	return io.NopCloser(reader), nil
}

func (el *EncryptedLayer) Uncompressed() (io.ReadCloser, error) {
	// Uncompressed data is not available for encrypted layers
	return nil, fmt.Errorf("uncompressed data is not available for encrypted layers")
}

func (el *EncryptedLayer) Size() (int64, error) {
	// Size of the encrypted layer is not the same as the original layer
	return 0, fmt.Errorf("size is not available for encrypted layers")
}

func (el *EncryptedLayer) MediaType() (types.MediaType, error) {
	// MediaType of the encrypted layer is not the same as the original layer
	return "", fmt.Errorf("media type is not available for encrypted layers")
}
