package jwt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	manager, err := ProvideTokenManagerE(cfg)
	if err != nil {
		panic(err)
	}
	return manager
}

// ProvideTokenManagerE creates a TokenManager and returns configuration errors.
func ProvideTokenManagerE(cfg *Config) (*TokenManager, error) {
	if cfg == nil || cfg.Secret == "" {
		return NewTokenManager(""), nil
	}

	tokenConfig := &TokenConfig{}

	if cfg.AccessTokenExpiry != "" {
		duration, err := parseConfigDuration(cfg.AccessTokenExpiry)
		if err != nil {
			return nil, fmt.Errorf("access token expiry: %w", err)
		}
		tokenConfig.AccessTokenExpiry = duration
	}
	if cfg.RefreshTokenExpiry != "" {
		duration, err := parseConfigDuration(cfg.RefreshTokenExpiry)
		if err != nil {
			return nil, fmt.Errorf("refresh token expiry: %w", err)
		}
		tokenConfig.RefreshTokenExpiry = duration
	}
	if cfg.RegisterTokenExpiry != "" {
		duration, err := parseConfigDuration(cfg.RegisterTokenExpiry)
		if err != nil {
			return nil, fmt.Errorf("register token expiry: %w", err)
		}
		tokenConfig.RegisterTokenExpiry = duration
	}

	return NewTokenManager(cfg.Secret, tokenConfig), nil
}

// ProvideTokenManagerFromSecret creates a TokenManager directly from a secret string.
// This is a convenience provider for simple use cases.
func ProvideTokenManagerFromSecret(secret string) *TokenManager {
	return NewTokenManager(secret)
}

func parseConfigDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("duration is empty")
	}

	unit := value[len(value)-1]
	if unit == 'd' || unit == 'D' || unit == 'w' || unit == 'W' {
		amount, err := strconv.ParseInt(value[:len(value)-1], 10, 64)
		if err != nil || amount <= 0 {
			return 0, fmt.Errorf("invalid duration %q", value)
		}
		if unit == 'd' || unit == 'D' {
			return time.Duration(amount) * 24 * time.Hour, nil
		}
		return time.Duration(amount) * 7 * 24 * time.Hour, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero")
	}
	return duration, nil
}
