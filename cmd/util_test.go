package main

import (
	"testing"

	"github.com/nicholasjackson/kapsule/crypto/keyproviders"
	"github.com/stretchr/testify/require"
)

func TestProviderReturnsErrorWithFileParamsAndVaultParams(t *testing.T) {
	_, err := getKeyProvider("public", "private", "/keys/", "kapsule", "token", "addr", "namespace")
	require.Error(t, err)
}

func TestProviderReturnsErrorWithMissingVaultParams(t *testing.T) {
	_, err := getKeyProvider("", "", "/keys/", "kapsule", "", "", "")
	require.Error(t, err)

	_, err = getKeyProvider("", "", "/keys/", "", "addr", "", "")
	require.Error(t, err)
}

func TestProviderReturnsFileProvider(t *testing.T) {
	kp, err := getKeyProvider("public.key", "", "", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, kp)
	require.IsType(t, &keyproviders.File{}, kp)
}
func TestProviderReturnsVaultProvider(t *testing.T) {
	kp, err := getKeyProvider("", "", "transit", "kapsule", "addr", "root", "")
	require.NoError(t, err)
	require.NotNil(t, kp)
	require.IsType(t, &keyproviders.Vault{}, kp)
}
