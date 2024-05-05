package reader

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/nicholasjackson/kapsule/builder"
	"github.com/nicholasjackson/kapsule/crypto/keyproviders"
	"github.com/nicholasjackson/kapsule/writer"
	"github.com/stretchr/testify/require"
)

func setupRegistry(t *testing.T, ref string) (*OCIRegistry, *log.Logger) {
	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	// create a builder and push to a registry
	kp := keyproviders.NewFile("../test_fixtures/testmodel/public.key", "../test_fixtures/testmodel/private.key")
	b := builder.NewBuilder()
	w := writer.NewOCIRegistry(l, kp, "admin", "password", true)

	// build the image
	i, err := b.Build("../test_fixtures/testmodel/modelfile", "../test_fixtures/testmodel")
	require.NoError(t, err)

	// push the image
	err = w.Write(i, ref, false, false)
	require.NoError(t, err)

	return NewOCIRegistry(l, "admin", "password", true), l
}

func TestACCPullFromRegistry(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	ref := "auth.container.local.jmpd.in:5001/testmodel:plain"

	r, _ := setupRegistry(t, ref)

	i, err := r.Pull(ref)
	require.NoError(t, err)
	require.NotNil(t, i)
}
