package main

import (
	"fmt"

	"github.com/nicholasjackson/kapsule/crypto/keyproviders"
)

func getKeyProvider(
	publicKey,
	privateKey,
	encryptionVaultPath,
	encryptionVaultKey,
	encryptionVaultAuthToken,
	encryptionVaultAuthAddr,
	encryptionVaultAuthNamespace string) (keyproviders.Provider, error) {

	// ensure that the user has not specified both a file based key and a vault key
	if (publicKey != "" || privateKey != "") && encryptionVaultKey != "" {
		return nil, fmt.Errorf("cannot specify both encryption key and vault key")
	}

	if encryptionVaultKey != "" && (encryptionVaultAuthAddr == "" || encryptionVaultAuthToken == "" || encryptionVaultPath == "") {
		return nil, fmt.Errorf("you must specify the vault address, token, transit path and keyname when using a vault key")
	}

	if publicKey != "" || privateKey != "" {
		return keyproviders.NewFile(publicKey, privateKey), nil
	}

	if encryptionVaultPath != "" && encryptionVaultKey != "" && encryptionVaultAuthToken != "" && encryptionVaultAuthAddr != "" {
		return keyproviders.NewVault(nil, encryptionVaultPath, encryptionVaultKey, "", encryptionVaultAuthToken, encryptionVaultAuthAddr, encryptionVaultAuthNamespace), nil
	}

	return nil, fmt.Errorf("you must specify either a file based key or a vault key")
}
