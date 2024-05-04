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

	if encryptionVaultKey != "" && encryptionVaultAuthToken != "" && (encryptionVaultAuthAddr == "" || encryptionVaultAuthNamespace == "") {
		return nil, fmt.Errorf("you must specify both the vault address and token when using a vault key")
	}

	return nil, nil
}
