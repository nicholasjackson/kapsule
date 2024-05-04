package main

import (
	"fmt"

	"github.com/nicholasjackson/kapsule/crypto"
)

func getKeyProvider(
	publicKey,
	privateKey,
	encryptionVaultKey,
	encryptionVaultAuthToken,
	encryptionVaultAuthAddr,
	encryptionVaultAuthNamespace string) (crypto.KeyProvider, error) {

	// ensure that the user has not specified both a file based key and a vault key
	if (publicKey != "" || privateKey != "") && encryptionVaultKey != "" {
		return nil, fmt.Errorf("cannot specify both encryption key and vault key")
	}

	if encryptionVaultKey != "" && (encryptionVaultAuthAddr == "" || encryptionVaultAuthToken == "") {
		return nil, fmt.Errorf("you must specify both the vault address and token when using a vault key")
	}

	if publicKey != "" || privateKey != "" {
		return crypto.NewKeyProviderFile(publicKey, privateKey), nil
	}

	return nil, nil
}
