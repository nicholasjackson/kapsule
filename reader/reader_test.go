package reader

import (
	"os"
	"testing"

	"github.com/nicholasjackson/kapsule/writer"
	"github.com/stretchr/testify/require"
)

func TestPullFromRegistry(t *testing.T) {
	r := &ReaderImpl{}

	i, err := r.PullFromRegistry("docker.io/nicholasjackson/llm_test:latest", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	require.NoError(t, err)
	require.NotNil(t, i)

	// try to write to a path
	d := t.TempDir()
	err = writer.WriteToPath(i, d)
	require.NoError(t, err)
}
