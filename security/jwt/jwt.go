package jwt

import (
	"time"

	jwtstd "github.com/golang-jwt/jwt/v5"
)

// Token expiration constants
const (
	DefaultAccessTokenExpire   = 2 * time.Hour
	DefaultRefreshTokenExpire  = 7 * 24 * time.Hour
	DefaultRegisterTokenExpire = time.Hour
)

// Error constants
const (
	ErrNeedTokenProvider = TokenError("token provider required")
	ErrInvalidToken      = TokenError("invalid token")
	ErrTokenExpired      = TokenError("token expired")
	ErrTokenParsing      = TokenError("token parsing error")
)

// TokenError represents JWT token related errors
type TokenError string

func (e TokenError) Error() string {
	return string(e)
}

// TokenManager handles JWT token operations
type TokenManager struct {
	secret string
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: secret}
}

// generateToken creates a JWT token with specified parameters
func (tm *TokenManager) generateToken(jti string, subject string, payload map[string]any, expiry time.Duration) (string, error) {
	if tm.secret == "" {
		return "", ErrNeedTokenProvider
	}

	now := time.Now()
	claims := jwtstd.MapClaims{
		"jti": jti,
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(expiry).Unix(),
	}

	if payload != nil && len(payload) > 0 {
		claims["payload"] = payload
	}

	token := jwtstd.NewWithClaims(jwtstd.SigningMethodHS256, claims)
	return token.SignedString([]byte(tm.secret))
}

// GenerateAccessToken generates an access token
func (tm *TokenManager) GenerateAccessToken(jti string, payload map[string]any) (string, error) {
	return tm.generateToken(jti, "access", payload, DefaultAccessTokenExpire)
}

// GenerateAccessTokenWithExpiry generates an access token with custom expiry
func (tm *TokenManager) GenerateAccessTokenWithExpiry(jti string, payload map[string]any, expiry time.Duration) (string, error) {
	return tm.generateToken(jti, "access", payload, expiry)
}

// GenerateRefreshToken generates a refresh token
func (tm *TokenManager) GenerateRefreshToken(jti string, payload map[string]any) (string, error) {
	return tm.generateToken(jti, "refresh", payload, DefaultRefreshTokenExpire)
}

// GenerateRefreshTokenWithExpiry generates a refresh token with custom expiry
func (tm *TokenManager) GenerateRefreshTokenWithExpiry(jti string, payload map[string]any, expiry time.Duration) (string, error) {
	return tm.generateToken(jti, "refresh", payload, expiry)
}

// GenerateRegisterToken generates a register token
func (tm *TokenManager) GenerateRegisterToken(jti string, payload map[string]any, subject string) (string, error) {
	return tm.generateToken(jti, subject, payload, DefaultRegisterTokenExpire)
}

// ValidateToken validates a JWT token and returns the parsed token
func (tm *TokenManager) ValidateToken(tokenString string) (*jwtstd.Token, error) {
	if tm.secret == "" {
		return nil, ErrNeedTokenProvider
	}

	token, err := jwtstd.Parse(tokenString, func(token *jwtstd.Token) (any, error) {
		if _, ok := token.Method.(*jwtstd.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(tm.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// DecodeToken decodes a JWT token and returns its claims
func (tm *TokenManager) DecodeToken(tokenString string) (map[string]any, error) {
	token, err := tm.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwtstd.MapClaims)
	if !ok {
		return nil, ErrTokenParsing
	}

	return claims, nil
}

// GetPayload extracts the payload from token claims
func (tm *TokenManager) GetPayload(tokenString string) (map[string]any, error) {
	claims, err := tm.DecodeToken(tokenString)
	if err != nil {
		return nil, err
	}

	payload, ok := claims["payload"].(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}

	return payload, nil
}

// IsTokenExpired checks if a token is expired
func (tm *TokenManager) IsTokenExpired(tokenString string) bool {
	claims, err := tm.DecodeToken(tokenString)
	if err != nil {
		return true
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return true
	}

	return time.Unix(int64(exp), 0).Before(time.Now())
}

// GetTokenExpiry returns the expiry time of a token
func (tm *TokenManager) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := tm.DecodeToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, ErrTokenParsing
	}

	return time.Unix(int64(exp), 0), nil
}

// RefreshTokenIfNeeded refreshes token if it's close to expiry
func (tm *TokenManager) RefreshTokenIfNeeded(tokenString string, refreshThreshold time.Duration) (string, bool, error) {
	claims, err := tm.DecodeToken(tokenString)
	if err != nil {
		return "", false, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", false, ErrTokenParsing
	}

	expiryTime := time.Unix(int64(exp), 0)
	if time.Until(expiryTime) > refreshThreshold {
		return tokenString, false, nil // No refresh needed
	}

	// Extract payload and regenerate token
	payload, ok := claims["payload"].(map[string]any)
	if !ok {
		payload = map[string]any{}
	}

	jti, _ := claims["jti"].(string)
	newToken, err := tm.GenerateAccessToken(jti, payload)
	return newToken, true, err
}
