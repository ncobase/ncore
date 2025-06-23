package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// StateManager manages OAuth state parameters
type StateManager struct {
	secret []byte
}

// NewStateManager creates a new state manager
func NewStateManager(secret string) *StateManager {
	return &StateManager{
		secret: []byte(secret),
	}
}

// GenerateState generates a secure state parameter
func (sm *StateManager) GenerateState(data *StateData) (string, error) {
	if data.Timestamp == 0 {
		data.Timestamp = time.Now().Unix()
	}
	if data.Nonce == "" {
		data.Nonce = sm.generateNonce()
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// In production, you should encrypt the state data
	return base64.URLEncoding.EncodeToString(jsonData), nil
}

// ParseState parses and validates state parameter
func (sm *StateManager) ParseState(state string) (*StateData, error) {
	jsonData, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return nil, err
	}

	var data StateData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Validate timestamp (5 minutes expiry)
	if time.Now().Unix()-data.Timestamp > 300 {
		return nil, errors.New("state expired")
	}

	return &data, nil
}

// GeneratePKCE generates PKCE challenge data
func (sm *StateManager) GeneratePKCE() (*PKCEData, error) {
	// Generate code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	codeVerifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(verifierBytes)

	// Generate code challenge using S256
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil))

	return &PKCEData{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		Method:        "S256",
	}, nil
}

// generateNonce generates a random nonce
func (sm *StateManager) generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
