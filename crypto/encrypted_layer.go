package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	"github.com/containers/ocicrypt/utils"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/types"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type EncryptedLayer struct {
	layer         v1.Layer
	key           []byte
	encryptConfig *config.EncryptConfig
	annotations   map[string]string
	hash          string
	size          int
	done          bool
}

func NewEncryptedLayer(l v1.Layer, key []byte) (*EncryptedLayer, error) {
	// check the key is a public key
	pub := utils.IsPublicKey(key)
	if !pub {
		return nil, fmt.Errorf("gven key is not a public key")
	}

	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{
			"pubkeys": {key},
		},
		DecryptConfig: config.DecryptConfig{},
	}

	return &EncryptedLayer{
		layer:         l,
		key:           key,
		encryptConfig: ec,
	}, nil
}

func (el *EncryptedLayer) Digest() (v1.Hash, error) {
	if el.done {
		return v1.NewHash(fmt.Sprintf("sha256:%s", el.hash))
	}

	// if we have not yet completed the encryption process we cannot get the digest
	// so we need to return a special error so that the writer package
	// will know to fetch this info later
	return v1.Hash{}, stream.ErrNotComputed
}

func (el *EncryptedLayer) Annotations() (map[string]string, error) {
	if el.done {
		// return the annotations that are created by the encryption process
		// this can only be called after the layer has been consumed
		return el.annotations, nil
	}

	return map[string]string{}, stream.ErrNotComputed
}

func (el *EncryptedLayer) DiffID() (v1.Hash, error) {
	// DiffID of the encrypted layer is not the same as the original layer
	return el.Digest()
}

func (el *EncryptedLayer) Compressed() (io.ReadCloser, error) {
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
	des := ocispec.Descriptor{}
	er, efin, err := ocicrypt.EncryptLayer(el.encryptConfig, r, des)
	if err != nil {
		return nil, err
	}

	var wr *wrappedReader

	// create the finalizer that will be called once the reader is closed
	// this will allow us to set the annotations, digest and size
	fin := func() error {
		// get the annotations from the encrypted layer, this information
		// is needed in order to decrypt the data later
		annot, err := efin()
		if err != nil {
			return fmt.Errorf("unable to get annotations from encrypted reader: %w", err)
		}

		// set the annotations on the encrypted layer
		el.annotations = annot

		// get the hash of the encrypted data
		h, err := wr.Hash()
		if err != nil {
			return fmt.Errorf("unable to get hash from wrapped reader: %w", err)
		}

		el.hash = h

		// get the size of the encrypted data
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
	wr = newWrappedReader(er, fin)

	// return wrapped reader that will return the encrypted data
	return wr, nil
}

func (el *EncryptedLayer) Uncompressed() (io.ReadCloser, error) {
	// Uncompressed data is not available for encrypted layers
	return nil, fmt.Errorf("uncompressed data is not available for encrypted layers")
}

func (el *EncryptedLayer) Size() (int64, error) {
	if el.size == 0 {
		return 0, stream.ErrNotComputed
	}

	return int64(el.size), nil
}

func (el *EncryptedLayer) MediaType() (types.MediaType, error) {
	// MediaType of the encrypted layer is the same as the original +enc
	mt, err := el.layer.MediaType()
	if err != nil {
		return "", err
	}

	return types.MediaType(mt + "+enc"), nil
}
