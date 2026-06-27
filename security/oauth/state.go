package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const defaultStateTTL = 5 * time.Minute

// StateManager manages OAuth state parameters
type StateManager struct {
	secret []byte
	ttl    time.Duration
}

// NewStateManager creates a new state manager
func NewStateManager(secret string) *StateManager {
	return &StateManager{
		secret: []byte(secret),
		ttl:    defaultStateTTL,
	}
}

// GenerateState generates a secure state parameter
func (sm *StateManager) GenerateState(data *StateData) (string, error) {
	if sm == nil || len(sm.secret) == 0 {
		return "", ErrInvalidConfiguration
	}
	if data == nil {
		return "", ErrInvalidState
	}

	if data.Timestamp == 0 {
		data.Timestamp = time.Now().Unix()
	}
	if data.Nonce == "" {
		nonce, err := sm.generateNonce()
		if err != nil {
			return "", err
		}
		data.Nonce = nonce
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString(jsonData)
	signature := sm.sign(payload)
	return payload + "." + signature, nil
}

// ParseState parses and validates state parameter
func (sm *StateManager) ParseState(state string) (*StateData, error) {
	if sm == nil || len(sm.secret) == 0 {
		return nil, ErrInvalidConfiguration
	}

	parts := strings.Split(state, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, ErrInvalidState
	}

	expectedSignature := sm.sign(parts[0])
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidState
	}
	expectedBytes, err := base64.RawURLEncoding.DecodeString(expectedSignature)
	if err != nil {
		return nil, ErrInvalidState
	}
	if !hmac.Equal(providedSignature, expectedBytes) {
		return nil, ErrInvalidState
	}

	jsonData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	var data StateData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	if data.Timestamp <= 0 {
		return nil, ErrInvalidState
	}

	now := time.Now()
	issuedAt := time.Unix(data.Timestamp, 0)
	if issuedAt.After(now.Add(time.Minute)) {
		return nil, ErrInvalidState
	}
	if now.Sub(issuedAt) > sm.ttl {
		return nil, ErrStateExpired
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

func (sm *StateManager) sign(payload string) string {
	mac := hmac.New(sha256.New, sm.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// generateNonce generates a random nonce.
func (sm *StateManager) generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate OAuth nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
