package main

import (
	"testing"

	"github.com/nicholasjackson/kapsule/crypto"
	"github.com/stretchr/testify/require"
)

func TestProviderReturnsErrorWithFileParamsAndVaultParams(t *testing.T) {
	_, err := getKeyProvider("public", "private", "/keys/", "token", "addr", "namespace")
	require.Error(t, err)
}

func TestProviderReturnsErrorWithMissingVaultParams(t *testing.T) {
	_, err := getKeyProvider("", "", "/keys/", "token", "", "")
	require.Error(t, err)

	_, err = getKeyProvider("", "", "/keys/", "", "addr", "")
	require.Error(t, err)
}

func TestProviderReturnsFileProviderWithPublicKey(t *testing.T) {
	kp, err := getKeyProvider("public.key", "", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, kp)
	require.IsType(t, &crypto.KeyProviderFile{}, kp)
}
