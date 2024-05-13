package crypto

import (
	"bytes"
	"encoding/hex"
	"hash"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrappedReaderReadsBase(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("hello world")))

	wr := newWrappedReader(r, func(h hash.Hash, count int64) error { return nil })

	d, err := io.ReadAll(wr)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(d))
}

func TestWrappedReaderHashesBaseStream(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("hello world")))

	var returnedHash hash.Hash
	wr := newWrappedReader(r, func(h hash.Hash, count int64) error {
		returnedHash = h
		return nil
	})

	_, err := io.ReadAll(wr)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", hex.EncodeToString(returnedHash.Sum(nil)))
}

func TestWrappedReaderCalculatesLength(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("hello world")))

	var s int64
	wr := newWrappedReader(r, func(h hash.Hash, count int64) error {
		s = count
		return nil
	})

	_, err := io.ReadAll(wr)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, int64(11), s)
}
