package oauth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleIssuer  = "https://appleid.apple.com"
	appleJWKSURL = "https://appleid.apple.com/auth/keys"
)

// ProviderSpecificClient handles provider-specific OAuth implementations
type ProviderSpecificClient struct {
	*Client

	appleKeysMu      sync.RWMutex
	appleKeys        map[string]*rsa.PublicKey
	appleKeysExpires time.Time
}

// NewProviderSpecificClient creates a client with provider-specific implementations
func NewProviderSpecificClient(config *Config) *ProviderSpecificClient {
	client, err := NewProviderSpecificClientE(config)
	if err != nil {
		panic(err)
	}
	return client
}

// NewProviderSpecificClientE creates a provider-specific client and returns
// configuration errors instead of silently accepting unsafe defaults.
func NewProviderSpecificClientE(config *Config) (*ProviderSpecificClient, error) {
	client, err := NewClientE(config)
	if err != nil {
		return nil, err
	}
	return &ProviderSpecificClient{
		Client:    client,
		appleKeys: make(map[string]*rsa.PublicKey),
	}, nil
}

// HandleAppleAuth handles Apple Sign In specific logic
func (c *ProviderSpecificClient) HandleAppleAuth(ctx context.Context, code, idToken string) (*Profile, *TokenResponse, error) {
	tokenResp, err := c.ExchangeCodeForToken(ctx, ProviderApple, code)
	if err != nil {
		return nil, nil, err
	}
	if idToken == "" {
		idToken = tokenResp.IDToken
	}
	if idToken == "" {
		return nil, nil, fmt.Errorf("%w: Apple id_token is required", ErrInvalidToken)
	}

	profile, err := c.parseAppleIDToken(ctx, idToken)
	if err != nil {
		return nil, nil, err
	}

	return profile, tokenResp, nil
}

// parseAppleIDToken parses Apple ID token to extract user profile
func (c *ProviderSpecificClient) parseAppleIDToken(ctx context.Context, idToken string) (*Profile, error) {
	config, exists := c.config.Providers[string(ProviderApple)]
	if !exists || !config.Enabled || config.ClientID == "" {
		return nil, fmt.Errorf("%w: Apple provider is not configured", ErrInvalidConfiguration)
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		idToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodRS256 {
				return nil, ErrInvalidToken
			}
			kid, ok := token.Header["kid"].(string)
			if !ok || kid == "" {
				return nil, ErrInvalidToken
			}
			return c.applePublicKey(ctx, kid)
		},
		jwt.WithIssuer(appleIssuer),
		jwt.WithAudience(config.ClientID),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	profile := &Profile{
		Provider: string(ProviderApple),
		ID:       getString(claims, "sub"),
		Email:    getString(claims, "email"),
		Verified: getBool(claims, "email_verified"),
	}
	if profile.ID == "" {
		return nil, fmt.Errorf("%w: Apple id_token missing subject", ErrInvalidToken)
	}

	if name, ok := claims["name"].(map[string]any); ok {
		firstName := getString(name, "firstName")
		lastName := getString(name, "lastName")
		if firstName != "" || lastName != "" {
			profile.Name = fmt.Sprintf("%s %s", firstName, lastName)
		}
	}

	return profile, nil
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	KeyType string `json:"kty"`
	KeyID   string `json:"kid"`
	Use     string `json:"use"`
	Alg     string `json:"alg"`
	N       string `json:"n"`
	E       string `json:"e"`
}

func (c *ProviderSpecificClient) applePublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.appleKeysMu.RLock()
	if key, ok := c.appleKeys[kid]; ok && time.Now().Before(c.appleKeysExpires) {
		c.appleKeysMu.RUnlock()
		return key, nil
	}
	c.appleKeysMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleJWKSURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Apple JWKS: status %d", resp.StatusCode)
	}

	var keys appleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	parsed := make(map[string]*rsa.PublicKey, len(keys.Keys))
	for _, key := range keys.Keys {
		if key.KeyType != "RSA" || key.N == "" || key.E == "" {
			continue
		}
		publicKey, err := rsaPublicKeyFromJWK(key)
		if err != nil {
			continue
		}
		parsed[key.KeyID] = publicKey
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("%w: Apple JWKS did not contain RSA keys", ErrInvalidToken)
	}

	c.appleKeysMu.Lock()
	c.appleKeys = parsed
	c.appleKeysExpires = time.Now().Add(time.Hour)
	key := c.appleKeys[kid]
	c.appleKeysMu.Unlock()
	if key == nil {
		return nil, fmt.Errorf("%w: Apple signing key %q not found", ErrInvalidToken, kid)
	}
	return key, nil
}

func rsaPublicKeyFromJWK(key appleJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}
	exponent := 0
	for _, b := range eBytes {
		exponent = exponent<<8 + int(b)
	}
	if exponent == 0 {
		return nil, ErrInvalidToken
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: exponent,
	}, nil
}

// GenerateAppleClientSecret generates Apple client secret JWT
func (c *ProviderSpecificClient) GenerateAppleClientSecret(keyID, teamID, clientID, privateKeyPEM string) (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}

	// Create JWT claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 180).Unix(), // Apple allows up to 6 months
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	// Sign token
	tokenString, err := token.SignedString(rsaPrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// HandleTwitterOAuth2 handles Twitter OAuth 2.0 with PKCE
func (c *ProviderSpecificClient) HandleTwitterOAuth2(ctx context.Context, codeVerifier, code string) (*Profile, *TokenResponse, error) {
	// Exchange code for token with PKCE
	tokenResp, err := c.ExchangeCodeForToken(ctx, ProviderTwitter, code, codeVerifier)
	if err != nil {
		return nil, nil, err
	}

	// Get user profile
	profile, err := c.GetUserProfile(ctx, ProviderTwitter, tokenResp.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	return profile, tokenResp, nil
}
