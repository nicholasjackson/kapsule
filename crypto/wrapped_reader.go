package crypto

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

// wrappedReader is an io.ReaderCloser, that wraps an io.Reader, hashing
// and counting the size of the data read from the reader.
type wrappedReader struct {
	pr       io.Reader
	h        hash.Hash
	c        *countWriter
	original io.Reader
	finalize func() error
	done     bool
}

// newWrappedReader creates a new wrappedReader that wraps the provided reader.
func newWrappedReader(r io.Reader, finalize func() error) *wrappedReader {
	pr, pw := io.Pipe()
	h := crypto.SHA256.New()
	c := &countWriter{}

	er := &wrappedReader{
		pr:       pr,
		original: r,
		finalize: finalize,
		h:        h,
		c:        c,
	}

	go func() {
		// Copy blob into the gzip writer, which also hashes and counts the
		// size of the encrypted output, and hasher of the enctyped contents.
		// blocks until reader is closed
		_, _ = io.Copy(io.MultiWriter(h, c, pw), er.original)

		er.done = true
		pw.Close()
	}()

	return er
}

// Read implements the io.Reader interface.
func (er *wrappedReader) Read(b []byte) (int, error) { return er.pr.Read(b) }

// Close implements the io.Closer interface.
func (er *wrappedReader) Close() error {
	if er.finalize != nil {
		return er.finalize()
	}

	return nil
}

// Size returns the size of the data read from the reader,
// this method can only be called once the reader is closed.
func (er *wrappedReader) Size() (int, error) {
	return er.c.n, nil
}

// Hash returns the hash of the data read from the reader,
// this method can only be called once the reader is closed.
func (er *wrappedReader) Hash() (string, error) {
	if !er.done {
		return "", fmt.Errorf("hash is not available until the reader is closed")
	}

	return hex.EncodeToString(er.h.Sum(nil)), nil
}

// countWriter counts bytes written to it.
type countWriter struct{ n int }

// Write implements the io.Writer interface.
func (c *countWriter) Write(p []byte) (int, error) {
	c.n += len(p)
	return len(p), nil
}
