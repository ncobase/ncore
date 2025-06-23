package oauth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ProviderSpecificClient handles provider-specific OAuth implementations
type ProviderSpecificClient struct {
	*Client
}

// NewProviderSpecificClient creates a client with provider-specific implementations
func NewProviderSpecificClient(config *Config) *ProviderSpecificClient {
	return &ProviderSpecificClient{
		Client: NewClient(config),
	}
}

// HandleAppleAuth handles Apple Sign In specific logic
func (c *ProviderSpecificClient) HandleAppleAuth(ctx context.Context, code, idToken string) (*Profile, *TokenResponse, error) {
	// Exchange code for token
	tokenResp, err := c.ExchangeCodeForToken(ctx, ProviderApple, code)
	if err != nil {
		return nil, nil, err
	}

	// Parse ID token for user info (Apple doesn't provide userinfo endpoint)
	profile, err := c.parseAppleIDToken(idToken)
	if err != nil {
		return nil, nil, err
	}

	return profile, tokenResp, nil
}

// parseAppleIDToken parses Apple ID token to extract user profile
func (c *ProviderSpecificClient) parseAppleIDToken(idToken string) (*Profile, error) {
	// Parse JWT token without verification (you should verify in production)
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	profile := &Profile{
		Provider: string(ProviderApple),
		ID:       getString(claims, "sub"),
		Email:    getString(claims, "email"),
		Verified: getBool(claims, "email_verified"),
	}

	// Extract name if present (only on first auth)
	if name, ok := claims["name"].(map[string]interface{}); ok {
		firstName := getString(name, "firstName")
		lastName := getString(name, "lastName")
		if firstName != "" || lastName != "" {
			profile.Name = fmt.Sprintf("%s %s", firstName, lastName)
		}
	}

	return profile, nil
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
