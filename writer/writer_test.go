package writer

import (
	"os"
	"path"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/nicholasjackson/kapsule/builder"
	"github.com/nicholasjackson/kapsule/reader"
	"github.com/stretchr/testify/require"
)

func setupWriter(t *testing.T) (image v1.Image, output string) {
	td := t.TempDir()
	o := path.Join(td, "output")

	os.MkdirAll(o, os.ModePerm)

	b := builder.NewBuilder()
	i, err := b.Build("../test_fixtures/testmodel/modelfile", "../test_fixtures/testmodel/")
	require.NoError(t, err)

	return i, o
}

func TestPullFromRegistryAndWritesDecryptedToPath(t *testing.T) {
	r := &reader.ReaderImpl{}
	i, err := r.PullFromRegistry("docker.io/nicholasjackson/mistral:encrypted", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	require.NoError(t, err)

	err = WriteToPath(i, "../output", "", "../test_fixtures/keys/private.key", true)
	require.NoError(t, err)
}

func TestACCWritesEncryptedToPath(t *testing.T) {
	//if os.Getenv("TEST_ACC") != "1" {
	//	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//}

	//td := t.TempDir()

	i, _ := setupWriter(t)

	err := WriteToPath(i, "../output", "../test_fixtures/keys/public.key", "", false)
	require.NoError(t, err)
}

func TestACCWritesPlainToPath(t *testing.T) {
	//if os.Getenv("TEST_ACC") != "1" {
	//	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//}

	//td := t.TempDir()

	i, _ := setupWriter(t)

	err := WriteToPath(i, "../output", "", "", true)
	require.NoError(t, err)
}

func TestACCWritesToRemoteRegistry(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	i, _ := setupWriter(t)

	err := PushToRegistry("docker.io/nicholasjackson/llm_test:latest", i, os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"), "")
	require.NoError(t, err)
}

func TestACCWriteToOllamaFormat(t *testing.T) {
	//if os.Getenv("TEST_ACC") != "1" {
	//	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//}

	i, _ := setupWriter(t)

	err := WriteToOllama(i, "kapsule.io/nicholasjackson/mistral:tune", "/home/nicj/go/src/github.com/nicholasjackson/demo-vault-securing-llm/cache/olama/models")
	require.NoError(t, err)
}
