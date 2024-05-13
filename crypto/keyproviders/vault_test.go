package keyproviders

import (
	"os"
	"testing"

	"github.com/containers/ocicrypt/utils"
	"github.com/nicholasjackson/kapsule/testutils"
	"github.com/stretchr/testify/require"
)

func setupVault(t *testing.T, key string) *Vault {
	l := testutils.CreateTestLogger(t)

	v := NewVault(l, "transit", key, "latest", "root", "http://vault.container.local.jmpd.in:8200", "")
	return v
}

func TestACCVaultReturnsErrorWhenNoPublicKey(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	kp := setupVault(t, "notfound")

	_, err := kp.PublicKey()
	require.Error(t, err)
}

func TestACCVaultReturnsPublicKey(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	kp := setupVault(t, "kapsule")

	k, err := kp.PublicKey()
	require.NoError(t, err)

	// check is valid public key
	require.True(t, utils.IsPublicKey(k), "Invalid public key")
}

func TestACCVaultReturnsErrorWhenNoPrivateKey(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	kp := setupVault(t, "notfound")

	_, err := kp.PrivateKey()
	require.Error(t, err)
}

func TestACCVaultReturnsPrivateKey(t *testing.T) {
	if os.Getenv("TEST_ACC") != "1" {
		t.Skip("Skipping test as Env var TEST_ACC is not set")
	}

	kp := setupVault(t, "kapsule")

	k, err := kp.PrivateKey()
	require.NoError(t, err)

	// check is valid public key
	isPK, _ := utils.IsPrivateKey(k, nil)
	require.True(t, isPK, "Invalid private key")
}
