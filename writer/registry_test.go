package writer

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/nicholasjackson/kapsule/builder"
	"github.com/nicholasjackson/kapsule/crypto/keyproviders"
	"github.com/stretchr/testify/require"
)

func setupRegistry(t *testing.T, ref string) (*OCIRegistry, *log.Logger, v1.Image) {
	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	// create a builder and push to a registry
	kp := keyproviders.NewFile("../test_fixtures/keys/public.key", "../test_fixtures/keys/private.key")
	b := builder.NewBuilder()
	w := NewOCIRegistry(l, kp, "admin", "password", true)

	// build the image
	i, err := b.Build("../test_fixtures/testmodel/modelfile", "../test_fixtures/testmodel")
	require.NoError(t, err)

	return w, l, i
}

func TestACCPushToRegistry(t *testing.T) {
	//if os.Getenv("TEST_ACC") != "1" {
	//	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//}

	ref := "auth.container.local.jmpd.in:5001/testmodel:plain"

	r, _, i := setupRegistry(t, ref)

	err := r.Write(i, ref, false, false)
	require.NoError(t, err)
}

func TestACCPushEncryptedToRegistry(t *testing.T) {
	//jif os.Getenv("TEST_ACC") != "1" {
	//j	t.Skip("Skipping test as Env var TEST_ACC is not set")
	//j}

	ref := "auth.container.local.jmpd.in:5001/testmodel:enc"

	r, _, i := setupRegistry(t, ref)

	err := r.WriteEncrypted(i, ref)
	require.NoError(t, err)
}
