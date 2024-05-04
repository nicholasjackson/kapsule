package crypto

type KeyProvider interface {
	// PublicKey returns the public key
	PublicKey() ([]byte, error)
	// PrivateKey returns the private key
	PrivateKey() ([]byte, error)
}
