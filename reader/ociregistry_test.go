package reader

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/nicholasjackson/kapsule/writer"
	"github.com/stretchr/testify/require"
)

func TestPullFromRegistry(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	r := NewOCIRegistry(l)

	i, err := r.Pull("docker.io/nicholasjackson/mistral:plain", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	require.NoError(t, err)
	require.NotNil(t, i)

	w := writer.NewPathWriter(l)
	// try to write to a path
	d := t.TempDir()
	err = w.Write(i, d, "", "", true)
	require.NoError(t, err)
}
