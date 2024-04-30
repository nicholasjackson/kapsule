package encryptor

import "github.com/containers/ocicrypt/config"

// Encryptor encrypts a container image layer
//
//go:generate mockery --name Encryptor
type Encryptor interface {
	Encrypt(foo string) (string, error)
	Decrypt(foo string) (string, error)
}

// EncryptorImpl is a concrete implementation of the Encryptor interface
type EncryptorImpl struct {
	encryptionConfig config.EncryptConfig
	decryptionConfig config.DecryptConfig
}

func NewEncryptor() Encryptor {
	return &EncryptorImpl{}
}

func (b *EncryptorImpl) Encrypt(foo string) (string, error) {
	return "", nil
}

func (b *EncryptorImpl) Decrypt(foo string) (string, error) {
	return "", nil
}
