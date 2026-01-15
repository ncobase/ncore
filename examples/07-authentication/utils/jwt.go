package auth

import (
	"fmt"
	"time"

	securityjwt "github.com/ncobase/ncore/security/jwt"
)

// Claims represents JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func GenerateJWT(userID, role, secret string, ttl time.Duration) (string, error) {
	manager := securityjwt.NewTokenManager(secret, &securityjwt.TokenConfig{AccessTokenExpiry: ttl})
	payload := map[string]any{
		"user_id": userID,
		"role":    role,
	}
	return manager.GenerateAccessToken(userID, payload, &securityjwt.TokenConfig{Expiry: ttl})
}

// ValidateJWT validates a JWT token and returns claims.
func ValidateJWT(tokenString, secret string) (map[string]any, error) {
	manager := securityjwt.NewTokenManager(secret)
	claims, err := manager.DecodeToken(tokenString)
	if err != nil {
		return nil, err
	}

	if !securityjwt.IsAccessToken(claims) {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func ParseJWT(tokenString, secret string) (*Claims, error) {
	manager := securityjwt.NewTokenManager(secret)
	claims, err := manager.DecodeToken(tokenString)
	if err != nil {
		return nil, err
	}

	payload := securityjwt.GetPayload(claims)
	return &Claims{
		UserID: securityjwt.GetString(payload, "user_id"),
		Role:   securityjwt.GetString(payload, "role"),
	}, nil
}
