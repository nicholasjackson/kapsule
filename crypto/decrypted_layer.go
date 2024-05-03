package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	"github.com/containers/ocicrypt/utils"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DecryptedLayer struct {
	layer         v1.Layer
	key           []byte
	decryptConfig *config.DecryptConfig
	annotations   map[string]string
	hash          string
	size          int
	done          bool
}

func NewDecryptedLayer(l v1.Layer, key []byte, annotations map[string]string) (*DecryptedLayer, error) {
	// check the key is a public key
	priv, _ := utils.IsPrivateKey(key, nil)
	if !priv {
		return nil, fmt.Errorf("gven key is not a private key")
	}

	dc := &config.DecryptConfig{
		Parameters: map[string][][]byte{
			"privkeys":           [][]byte{key},
			"privkeys-passwords": [][]byte{[]byte("")},
		},
	}

	return &DecryptedLayer{
		layer:         l,
		key:           key,
		decryptConfig: dc,
		annotations:   annotations,
	}, nil
}

func (el *DecryptedLayer) Digest() (v1.Hash, error) {
	if el.done {
		return v1.NewHash(fmt.Sprintf("sha256:%s", el.hash))
	}

	// if we have not yet completed the encryption process we cannot get the digest
	// so we need to return a special error so that the writer package
	// will know to fetch this info later
	return v1.Hash{}, stream.ErrNotComputed
}

func (el *DecryptedLayer) Annotations() (map[string]string, error) {
	if el.done {
		// return the annotations that are created by the encryption process
		// this can only be called after the layer has been consumed
		return el.annotations, nil
	}

	return map[string]string{}, stream.ErrNotComputed
}

func (el *DecryptedLayer) DiffID() (v1.Hash, error) {
	// DiffID of the encrypted layer is not the same as the original layer
	return el.Digest()
}

func (el *DecryptedLayer) Compressed() (io.ReadCloser, error) {
	// get the compressed layer reader, this will be passed to the encryptor
	r, err := el.layer.Compressed()
	if errors.Is(err, stream.ErrConsumed) {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to get compressed stream: %s", err)
	}

	// consturct the layer encryptor, the comppressed reader passed to the encryptor
	// will compresse the base stream before it is encrypted
	des := ocispec.Descriptor{
		Annotations: el.annotations,
	}

	dr, _, err := ocicrypt.DecryptLayer(el.decryptConfig, r, des, false)
	if err != nil {
		return nil, err
	}

	var wr *wrappedReader

	// create the finalizer that will be called once the reader is closed
	// this will allow us to set the annotations, digest and size
	fin := func() error {
		// get the hash of the decrypted data
		h, err := wr.Hash()
		if err != nil {
			return fmt.Errorf("unable to get hash from wrapped reader: %w", err)
		}

		el.hash = h

		// get the size of the decrypted data
		s, err := wr.Size()
		if err != nil {
			return fmt.Errorf("unable to get size from wrapped reader: %w", err)
		}

		el.size = s

		// mark the layer as done
		el.done = true

		return nil
	}

	// wrap the encrypted reader with a wrapped reader so that we
	// can get the hash and the size of the encrypted data
	wr = newWrappedReader(dr, fin)

	// return wrapped reader that will return the encrypted data
	return wr, nil
}

func (el *DecryptedLayer) Uncompressed() (io.ReadCloser, error) {
	// Uncompressed data is not available for encrypted layers
	return nil, fmt.Errorf("uncompressed data is not available for encrypted layers")
}

func (el *DecryptedLayer) Size() (int64, error) {
	if el.done {
		return int64(el.size), nil
	}

	// if we have not yet completed the encryption process we cannot get the digest
	// so we need to return a special error so that the writer package
	// will know to fetch this info later
	return -1, stream.ErrNotComputed
}

func (el *DecryptedLayer) MediaType() (types.MediaType, error) {
	// MediaType of the decrypted layer should have +enc stripped
	mt, err := el.layer.MediaType()
	if err != nil {
		return "", err
	}

	return types.MediaType(strings.TrimSuffix(string(mt), "+enc")), nil
}
