package crypto

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"testing"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	"github.com/containers/ocicrypt/utils"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/nicholasjackson/kapsule/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestEncryptedLayerEncryptsContent(t *testing.T) {

	templateLayer := stream.NewLayer(
		io.NopCloser(bytes.NewReader([]byte("hello world"))),
		stream.WithCompressionLevel(gzip.DefaultCompression),
		stream.WithMediaType(types.KAPSULE_MEDIA_TYPE_TEMPLATE),
	)

	// Load the public key from file
	pubKeyFile := "../test_fixtures/keys/public.key"
	pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
	require.NoError(t, err)

	privKeyFile := "../test_fixtures/keys/private.key"
	privKeyBytes, err := ioutil.ReadFile(privKeyFile)
	require.NoError(t, err)

	pub := utils.IsPublicKey(pubKeyBytes)
	require.True(t, pub)

	priv, _ := utils.IsPrivateKey(privKeyBytes, nil)
	require.True(t, priv)

	// Create an encryption config
	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{
			"pubkeys": [][]byte{pubKeyBytes},
		},
		DecryptConfig: config.DecryptConfig{
			Parameters: map[string][][]byte{
				"privkeys":           [][]byte{privKeyBytes},
				"privkeys-passwords": [][]byte{[]byte("")},
			},
		},
	}

	r, err := templateLayer.Compressed()
	require.NoError(t, err)

	// descriptor holds the annotation for the layer, etc
	des := ocispec.Descriptor{}

	er, fin, err := ocicrypt.EncryptLayer(ec, r, des)
	require.NoError(t, err)

	d, err := io.ReadAll(er)
	require.NoError(t, err)
	require.NotEqual(t, "hello world", string(d))

	annot, err := fin()
	require.NoError(t, err)
	require.NotEmpty(t, annot)

	des.Annotations = annot

	// Decrypt the layer
	crypt := bytes.NewReader(d)
	dr, _, err := ocicrypt.DecryptLayer(&ec.DecryptConfig, crypt, des, false)
	require.NoError(t, err)

	d, err = io.ReadAll(dr)
	require.NoError(t, err)
	require.NotEqual(t, "hello world", string(d))

	gz, err := gzip.NewReader(bytes.NewReader(d))
	require.NoError(t, err)

	d, err = io.ReadAll(gz)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(d))
}
