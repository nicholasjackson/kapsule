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

func setupPathFileKp(t *testing.T, ref string) (*PathWriter, *log.Logger, string, v1.Image) {
	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	// create a builder and push to a registry
	kp := keyproviders.NewFile("../test_fixtures/keys/public.key", "../test_fixtures/keys/private.key")
	b := builder.NewBuilder()

	// build the image
	i, err := b.Build("../test_fixtures/testmodel/modelfile", "../test_fixtures/testmodel")
	require.NoError(t, err)

	td := t.TempDir()

	return NewPathWriter(l, kp, td), l, td, i
}

func setupPathVaultKp(t *testing.T, ref string) (*PathWriter, *log.Logger, string, v1.Image) {
	l := log.New(os.Stdout)
	l.SetLevel(log.DebugLevel)

	// create a builder and push to a registry
	kp := keyproviders.NewVault(l, "transit", "kapsule", "latest", "root", "http://vault.container.local.jmpd.in:8200", "")
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

	pw, _, o, i := setupPathFileKp(t, "test")

	err := pw.Write(i, o, false, true)
	require.NoError(t, err)

	// check writen file exists

}

func TestACCPathWritesEncryptedWithFileKeyToPath(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	pw, _, o, i := setupPathFileKp(t, "test")

	err := pw.WriteEncrypted(i, o)
	require.NoError(t, err)

	// check writen file exists

}

func TestACCPathWritesEncryptedWithVaultKeyToPath(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	pw, _, o, i := setupPathVaultKp(t, "test")

	err := pw.WriteEncrypted(i, o)
	require.NoError(t, err)

	// check writen file exists

}
