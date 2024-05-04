package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReturnsErrorWhenNoPublicKey(t *testing.T) {
	kp := NewKeyProviderFile("./notexists.key", "")

	_, err := kp.PublicKey()

	require.Error(t, err)
}

func TestReturnsErrorWhenNotPublicKey(t *testing.T) {
	kp := NewKeyProviderFile("../test_fixtures/keys/private.key", "")

	_, err := kp.PublicKey()

	require.Error(t, err)
}

func TestReturnsPublicKey(t *testing.T) {
	kp := NewKeyProviderFile("../test_fixtures/keys/public.key", "")

	_, err := kp.PublicKey()

	require.NoError(t, err)
}

func TestReturnsErrorWhenNoPrivateKey(t *testing.T) {
	kp := NewKeyProviderFile("", "./notexists.key")

	_, err := kp.PrivateKey()

	require.Error(t, err)
}

func TestReturnsErrorWhenNotPrivateKey(t *testing.T) {
	kp := NewKeyProviderFile("", "../test_fixtures/keys/public.key")

	_, err := kp.PrivateKey()

	require.Error(t, err)
}

func TestReturnsPrivateKey(t *testing.T) {
	kp := NewKeyProviderFile("", "../test_fixtures/keys/private.key")

	_, err := kp.PrivateKey()

	require.NoError(t, err)
}
