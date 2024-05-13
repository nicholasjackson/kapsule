package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
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
	digest        v1.Hash
	size          int64
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
		return el.digest, nil
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
	// We need to get the DiffID from t he wrapped layer as this is the
	// unzipped unencrypted hash
	return el.layer.DiffID()
}

func (el *EncryptedLayer) Compressed() (io.ReadCloser, error) {
	if el.done {
		// with a standard layer we can only compress the data once and if the stream
		// has been written we should return an error. However with an ecnypted layer
		// the image is written twice, once with the encrypted data then the manifest
		// is updated with the encryption details.
		//
		// The path writer uses the Digest and the Size to check if a layer has been written
		// however it also calls Compressed to get the reader even if it does not use it
		// in this edge case we need to return an empty reader so that the path writer
		// does not fail second write
		return io.NopCloser(bytes.NewBuffer(nil)), nil
	}

	r, err := el.layer.Compressed()
	if err != nil {
		return nil, err
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
	fin := func(h hash.Hash, count int64) error {
		// get the annotations from the encrypted layer, this information
		// is needed in order to decrypt the data later
		annot, err := efin()
		if err != nil {
			return fmt.Errorf("unable to get annotations from encrypted reader: %w", err)
		}

		// set the annotations on the encrypted layer
		el.annotations = annot

		// get the hash of the encrypted data
		digest, err := v1.NewHash("sha256:" + hex.EncodeToString(h.Sum(nil)))
		if err != nil {
			return fmt.Errorf("unable to create hash from encrypted data: %w", err)
		}

		el.digest = digest
		el.size = count

		// mark the layer as done
		el.done = true

		return nil
	}

	// wrap the encrypted reader with a wrapped reader so that we
	// can get the hash and the size of the encrypted data
	wr = newWrappedReader(io.NopCloser(er), fin)

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

type wrappedR struct {
	pr       io.Reader
	finalize func() error
}

func (er *wrappedR) Read(b []byte) (int, error) {
	n, err := er.pr.Read(b)
	fmt.Println("er.pr.Read(b)", n)
	return n, err
}

// Close implements the io.Closer interface.
func (er *wrappedR) Close() error {
	if er.finalize != nil {
		return er.finalize()
	}

	return nil
}
