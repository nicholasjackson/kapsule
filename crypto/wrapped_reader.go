package crypto

import (
	"bufio"
	"crypto"
	"errors"
	"hash"
	"io"
	"os"
)

// wrappedReader is an io.ReaderCloser, that wraps an io.Reader, hashing
// and counting the size of the data read from the reader.
type wrappedReader struct {
	pr     *io.PipeReader
	pw     *io.PipeWriter
	closer func() error
	done   bool
}

// newWrappedReader creates a new wrappedReader that wraps the provided reader.
func newWrappedReader(r io.ReadCloser, finalize func(h hash.Hash, count int64) error) *wrappedReader {
	// we need to calculate the hash of the encrypted data
	// and the size of the encrypted data
	// the hash of the unencrypted unzipped data is collected from the base stream
	eh := crypto.SHA256.New()
	count := &countWriter{}

	// The encryption
	pr, pw := io.Pipe()

	// Write encrypted bytes to be read by the pipe.Reader, hashed by zh, and counted by count.
	mw := io.MultiWriter(pw, eh, count)

	// Buffer the output of the gzip writer so we don't have to wait on pr to keep writing.
	// 64K ought to be small enough for anybody.
	bw := bufio.NewWriterSize(mw, 2<<16)

	doneDigesting := make(chan struct{})

	er := &wrappedReader{
		pr: pr,
		pw: pw,
		closer: func() error {
			// Immediately close pw without error. There are three ways to get
			// here.
			//
			// 1. There was a copy error due from the underlying reader, in which
			//    case the error will not be overwritten.
			// 2. Copying from the underlying reader completed successfully.
			// 3. Close has been called before the underlying reader has been
			//    fully consumed. In this case pw must be closed in order to
			//    keep the flush of bw from blocking indefinitely.
			//
			// NOTE: pw.Close never returns an error. The signature is only to
			// implement io.Closer.
			_ = pw.Close()

			// Close the inner ReadCloser.
			//
			// NOTE: net/http will call close on success, so if we've already
			// closed the inner rc, it's not an error.
			if err := r.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				return err
			}

			<-doneDigesting
			return finalize(eh, count.n)
		},
	}

	go func() {
		// Copy blob into the gzip writer, which also hashes and counts the
		// size of the encrypted output, and hasher of the enctyped contents.
		// blocks until reader is closed
		_, copyErr := io.Copy(bw, r)

		if copyErr != nil {
			close(doneDigesting)
			pw.CloseWithError(copyErr)
			return
		}

		// Flush the buffer once all writes are complete to the gzip writer.
		if err := bw.Flush(); err != nil {
			close(doneDigesting)
			pw.CloseWithError(err)
			return
		}

		// Notify closer that digests are done being written.
		close(doneDigesting)

		// Close the compressed reader to calculate digest/diffID/size. This
		// will cause pr to return EOF which will cause readers of the
		// Compressed stream to finish reading.
		pw.CloseWithError(er.Close())
	}()

	return er
}

// Read implements the io.Reader interface.
func (er *wrappedReader) Read(b []byte) (int, error) { return er.pr.Read(b) }

// Close implements the io.Closer interface.
func (er *wrappedReader) Close() error { return er.closer() }

// countWriter counts bytes written to it.
type countWriter struct{ n int64 }

// Write implements the io.Writer interface.
func (c *countWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}
