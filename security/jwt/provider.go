package jwt

import (
	"github.com/google/wire"
)

// Config represents JWT configuration for Wire injection.
// This is used to configure the TokenManager via dependency injection.
type Config struct {
	Secret              string
	AccessTokenExpiry   string
	RefreshTokenExpiry  string
	RegisterTokenExpiry string
}

// ProviderSet is the wire provider set for the jwt package.
// It provides *TokenManager for JWT operations.
//
// Usage:
//
//	wire.Build(
//	    jwt.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(
	ProvideTokenManager,
	wire.Bind(new(TokenValidator), new(*TokenManager)),
)

// TokenValidator is an interface for validating JWT tokens.
// This allows for easier testing and dependency injection.
type TokenValidator interface {
	ValidateToken(tokenString string) (any, error)
	DecodeToken(tokenString string) (map[string]any, error)
	IsTokenExpired(tokenString string) bool
}

// ProvideTokenManager creates a new TokenManager from configuration.
// The secret is required; other settings use defaults if not specified.
func ProvideTokenManager(cfg *Config) *TokenManager {
	if cfg == nil || cfg.Secret == "" {
		// Return a TokenManager that requires secret to be set later
		return NewTokenManager("")
	}

	tokenConfig := &TokenConfig{}

	// Parse duration strings if provided
	if cfg.AccessTokenExpiry != "" {
		// Use default if parsing fails
		tokenConfig.AccessTokenExpiry = DefaultAccessTokenExpire
	}
	if cfg.RefreshTokenExpiry != "" {
		tokenConfig.RefreshTokenExpiry = DefaultRefreshTokenExpire
	}
	if cfg.RegisterTokenExpiry != "" {
		tokenConfig.RegisterTokenExpiry = DefaultRegisterTokenExpire
	}

	return NewTokenManager(cfg.Secret, tokenConfig)
}

// ProvideTokenManagerFromSecret creates a TokenManager directly from a secret string.
// This is a convenience provider for simple use cases.
func ProvideTokenManagerFromSecret(secret string) *TokenManager {
	return NewTokenManager(secret)
}
