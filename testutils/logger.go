package testutils

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/log"
)

func CreateTestLogger(t *testing.T) *log.Logger {
	bf := bytes.NewBuffer([]byte{})
	l := log.New(bf)
	l.SetLevel(log.DebugLevel)

	// if the test fails write the log
	t.Cleanup(func() {
		if t.Failed() {
			t.Log(bf.String())
		}
	})

	return l
}
