package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func generateAES256Key() ([]byte, error) {
	key := make([]byte, 32) // AES-256 key size
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func TestEncryptedLayerEncryptsContent(t *testing.T) {
	//templateLayer := stream.NewLayer(
	//	stream.WithCompressionLevel(gzip.NoCompression),
	//	stream.WithMediaType(types.KAPSULE_MEDIA_TYPE_TEMPLATE),
	//)

	key, err := generateAES256Key()
	require.NoError(t, err)

	l := NewEncryptedLayer(io.NopCloser(bytes.NewReader([]byte("template"))), key)

	r, err := l.Compressed()
	require.NoError(t, err)

	encrypted, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NotEqual(t, "template", string(encrypted))

	// the data should be encrypted
	// decrypt it
	dr, err := decryptReader(bytes.NewReader(encrypted), key)
	require.NoError(t, err)
	plain, err := io.ReadAll(dr)
	require.NoError(t, err)

	// check it is the same as original
	require.Equal(t, "template", string(plain))
}

func decryptReader(r io.Reader, key []byte) (io.Reader, error) {
	block, err := aes.NewCipher(key)
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
