package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderReturnsErrorWithFileParamsAndVaultParams(t *testing.T) {
	_, err := getKeyProvider("public", "private", "/keys/", "token", "addr", "namespace")
	require.Error(t, err)
}

func TestProviderReturnsErrorWithMissingVaultParams(t *testing.T) {
	_, err := getKeyProvider("", "", "/keys/", "token", "", "")
	require.Error(t, err)
}
