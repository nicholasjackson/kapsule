package keyproviders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

type Vault struct {
	logger        *log.Logger
	transitPath   string
	key           string
	version       string
	authToken     string
	authAddr      string
	authNamespace string
	publicKey     []byte
	privateKey    []byte
}

func NewVault(l *log.Logger, transitPath, key, version, authToken, authAddr, authNamespace string) *Vault {
	if version == "" {
		version = "latest"
	}

	return &Vault{
		logger:        l,
		transitPath:   strings.TrimSuffix(strings.TrimPrefix(transitPath, "/"), "/"),
		key:           key,
		version:       version,
		authToken:     authToken,
		authAddr:      strings.TrimSuffix(authAddr, "/"),
		authNamespace: authNamespace,
	}
}

func (kp *Vault) PublicKey() ([]byte, error) {
	if kp.publicKey != nil {
		return kp.publicKey, nil
	}

	k, err := kp.getVaultKey("public-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// set the key to cache
	kp.publicKey = k

	return kp.publicKey, nil
}

func (kp *Vault) PrivateKey() ([]byte, error) {
	if kp.privateKey != nil {
		return kp.privateKey, nil
	}

	k, err := kp.getVaultKey("encryption-key")
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// set the key to cache
	kp.privateKey = k

	return kp.privateKey, nil
}

func (kp *Vault) getVaultKey(keyType string) ([]byte, error) {
	path := fmt.Sprintf("%s/v1/%s/export/%s/%s/%s", kp.authAddr, kp.transitPath, keyType, kp.key, kp.version)
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	kp.logger.Debug("Sending request to Vault", "path", path)

	// add the token to the request
	r.Header.Add("X-Vault-Token", kp.authToken)

	// send the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// read the response
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get public key: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	kp.logger.Debug("Received response from Vault", "data", string(data))

	// unmarshal the response
	vaultResp := map[string]interface{}{}
	if err := json.Unmarshal(data, &vaultResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// get the public key
	keys := vaultResp["data"].(map[string]interface{})["keys"].(map[string]interface{})
	k, ok := keys["1"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	return []byte(k), nil
}
