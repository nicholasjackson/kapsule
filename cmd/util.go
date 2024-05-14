package main

import (
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/nicholasjackson/kapsule/crypto/keyproviders"
)

func getKeyProvider(
	l *log.Logger,
	publicKey,
	privateKey,
	encryptionVaultKey,
	encryptionVaultPath,
	encryptionVaultAuthToken,
	encryptionVaultAuthAddr,
	encryptionVaultAuthNamespace string) (keyproviders.Provider, error) {

	// if no key is specified return nil provider
	if publicKey == "" && privateKey == "" && encryptionVaultKey == "" {
		return &keyproviders.NullProvider{}, nil
	}

	// ensure that the user has not specified both a file based key and a vault key
	if (publicKey != "" || privateKey != "") && encryptionVaultKey != "" {
		return nil, fmt.Errorf("cannot specify both encryption key and vault key")
	}

	if encryptionVaultKey != "" && (encryptionVaultAuthAddr == "" || encryptionVaultPath == "") {
		return nil, fmt.Errorf("you must specify the vault address, transit path, and keyname when using a vault key")
	}

	if publicKey != "" || privateKey != "" {
		return keyproviders.NewFile(publicKey, privateKey), nil
	}

	if encryptionVaultPath != "" && encryptionVaultKey != "" && encryptionVaultAuthAddr != "" {
		return keyproviders.NewVault(l, encryptionVaultPath, encryptionVaultKey, "", encryptionVaultAuthToken, encryptionVaultAuthAddr, encryptionVaultAuthNamespace), nil
	}

	return nil, fmt.Errorf("you must specify either a file based key or a vault key")
}
