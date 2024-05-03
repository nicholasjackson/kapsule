package reader

import (
	"os"
	"testing"

	"github.com/nicholasjackson/kapsule/writer"
	"github.com/stretchr/testify/require"
)

func TestPullFromRegistry(t *testing.T) {
	//if os.Getenv("TEST_ACC") != "1" {
	//	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//}

	r := &ReaderImpl{}

	i, err := r.PullFromRegistry("docker.io/nicholasjackson/mistral:plain", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	require.NoError(t, err)
	require.NotNil(t, i)

	// try to write to a path
	//d := t.TempDir()
	d := "../output"
	err = writer.WriteToPath(i, d, "")
	require.NoError(t, err)
}
