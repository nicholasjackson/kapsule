package writer

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/nicholasjackson/kapsule/builder"
	"github.com/nicholasjackson/kapsule/crypto"
	"github.com/stretchr/testify/require"
)

func setupPath(t *testing.T, ref string) (*PathWriter, *log.Logger, string, v1.Image) {
	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	// create a builder and push to a registry
	kp := crypto.NewKeyProviderFile("../test_fixtures/testmodel/public.key", "../test_fixtures/testmodel/private.key")
	b := builder.NewBuilder()

	// build the image
	i, err := b.Build("../test_fixtures/testmodel/modelfile", "../test_fixtures/testmodel")
	require.NoError(t, err)

	td := t.TempDir()

	return NewPathWriter(l, kp, td), l, td, i
}

func TestACCPathWritesPlainToPath(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	pw, _, o, i := setupPath(t, "test")

	err := pw.Write(i, o, false, true)
	require.NoError(t, err)

	// check writen file exists
	
}
