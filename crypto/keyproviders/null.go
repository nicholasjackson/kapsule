package keyproviders

import "fmt"

type NullProvider struct{}

func (n *NullProvider) PublicKey() ([]byte, error) {
	return nil, fmt.Errorf("no public key available")
}

func (n *NullProvider) PrivateKey() ([]byte, error) {
	return nil, fmt.Errorf("no private key available")
}
