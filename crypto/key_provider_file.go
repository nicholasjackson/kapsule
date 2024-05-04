package crypto

import (
	"fmt"
	"os"

	"github.com/containers/ocicrypt/utils"
)

type KeyProviderFile struct {
	publicKeyPath  string
	privateKeyPath string
}

func NewKeyProviderFile(publicKeyPath, privateKeyPath string) *KeyProviderFile {
	return &KeyProviderFile{
		publicKeyPath:  publicKeyPath,
		privateKeyPath: privateKeyPath,
	}
}

func (kp *KeyProviderFile) PublicKey() ([]byte, error) {
	if kp.publicKeyPath == "" {
		return nil, fmt.Errorf("no public key configured")
	}

	d, err := os.ReadFile(kp.publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key: %s", err)
	}

	if !utils.IsPublicKey(d) {
		return nil, fmt.Errorf("%s is not a public key", kp.publicKeyPath)
	}

	return d, nil
}

func (kp *KeyProviderFile) PrivateKey() ([]byte, error) {
	if kp.privateKeyPath == "" {
		return nil, fmt.Errorf("no private key configured")
	}

	d, err := os.ReadFile(kp.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %s", err)
	}

	if ok, _ := utils.IsPrivateKey(d, []byte{}); !ok {
		return nil, fmt.Errorf("%s is not a private key", kp.privateKeyPath)
	}

	return d, nil
}
