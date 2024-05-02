package crypto

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrappedReaderReadsBase(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	wr := newWrappedReader(r, func() error { return nil })

	d, err := io.ReadAll(wr)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(d))
}

func TestWrappedReaderHashesBaseStream(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	wr := newWrappedReader(r, func() error { return nil })

	_, err := io.ReadAll(wr)
	require.NoError(t, err)

	h, err := wr.Hash()
	require.NoError(t, err)
	require.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", h)
}

func TestWrappedReaderCalculatesLength(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	wr := newWrappedReader(r, func() error { return nil })

	_, err := io.ReadAll(wr)
	require.NoError(t, err)

	s, err := wr.Size()
	require.NoError(t, err)
	require.Equal(t, 11, s)
}

func TestWrappedReaderReturnsErrorForSizeWhenStreamNotRead(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	wr := newWrappedReader(r, func() error { return nil })

	_, err := wr.Size()
	require.Error(t, err)
}

func TestWrappedReaderReturnsErrorForHashWhenStreamNotRead(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	wr := newWrappedReader(r, func() error { return nil })

	_, err := wr.Hash()
	require.Error(t, err)
}
