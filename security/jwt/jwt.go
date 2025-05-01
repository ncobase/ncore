package jwt

import (
	"time"

	jwtstd "github.com/golang-jwt/jwt/v5"
)

// TokenError represents JWT token related errors
type TokenError string

func (e TokenError) Error() string {
	return string(e)
}

const (
	DefaultAccessTokenExpire   = time.Hour * 24
	DefaultRegisterTokenExpire = time.Minute * 60
	DefaultRefreshTokenExpire  = time.Hour * 24 * 7

	ErrNeedTokenProvider = TokenError("cannot sign token without token provider")
	ErrInvalidToken      = TokenError("invalid token")
	ErrTokenParsing      = TokenError("token parsing error")
)

// TokenPayload represents the standard payload structure
type TokenPayload struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TenantID    string   `json:"tenant_id"`
	UserID      string   `json:"user_id"`
	IsAdmin     bool     `json:"is_admin"`
}

// Token represents the token body
type Token struct {
	JTI     string         `json:"jti"`
	Payload map[string]any `json:"payload"`
	Subject string         `json:"sub"`
	Expire  int64          `json:"exp"`
}

// TokenManager handles JWT token operations
type TokenManager struct {
	key string
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(key string) *TokenManager {
	return &TokenManager{key: key}
}

// validateKey validates the token key
func (jtm *TokenManager) validateKey() error {
	if jtm.key == "" {
		return ErrNeedTokenProvider
	}
	return nil
}

// generateToken generates a JWT token
func (jtm *TokenManager) generateToken(token *Token) (string, error) {
	if err := jtm.validateKey(); err != nil {
		return "", err
	}

	claims := jwtstd.MapClaims{
		"jti":     token.JTI,
		"sub":     token.Subject,
		"payload": token.Payload,
		"exp":     time.Now().Add(time.Millisecond * time.Duration(token.Expire)).Unix(),
	}

	t := jwtstd.NewWithClaims(jwtstd.SigningMethodHS256, claims)
	return t.SignedString([]byte(jtm.key))
}

// generateCustomToken generates a custom token with a specified expiration duration
func (jtm *TokenManager) generateCustomToken(jti string, payload map[string]any, subject string, expireDuration time.Duration) (string, error) {
	return jtm.generateToken(&Token{
		JTI:     jti,
		Payload: payload,
		Subject: subject,
		Expire:  expireDuration.Milliseconds(),
	})
}

// ensurePayloadDefaults ensures that the payload contains default values
func ensurePayloadDefaults(payload map[string]any) {
	defaults := map[string]any{
		"roles":       []string{},
		"permissions": []string{},
		"tenant_id":   "",
		"user_id":     "",
	}

	for key, defaultValue := range defaults {
		if _, exists := payload[key]; !exists {
			payload[key] = defaultValue
		}
	}
}

// GenerateAccessToken generates an access token with a default expiration of 24 hours
func (jtm *TokenManager) GenerateAccessToken(jti string, payload map[string]any, subject ...string) (string, error) {
	ensurePayloadDefaults(payload)
	return jtm.generateCustomToken(jti, payload, getSubject(subject, "access"), DefaultAccessTokenExpire)
}

// GenerateRegisterToken generates a register token with a default expiration of 60 minutes
func (jtm *TokenManager) GenerateRegisterToken(jti string, payload map[string]any, subject ...string) (string, error) {
	return jtm.generateCustomToken(jti, payload, getSubject(subject, "register"), DefaultRegisterTokenExpire)
}

// GenerateRefreshToken generates a refresh token with a default expiration of 7 days
func (jtm *TokenManager) GenerateRefreshToken(jti string, payload map[string]any, subject ...string) (string, error) {
	return jtm.generateCustomToken(jti, payload, getSubject(subject, "refresh"), DefaultRefreshTokenExpire)
}

// GenerateAccessTokenWithExpiry generates an access token with a custom expiration duration.
func (jtm *TokenManager) GenerateAccessTokenWithExpiry(jti string, payload map[string]any, expiry time.Duration, subject ...string) (string, error) {
	return jtm.generateCustomToken(jti, payload, getSubject(subject, "access"), expiry)
}

// GenerateRefreshTokenWithExpiry generates a refresh token with a custom expiration duration.
func (jtm *TokenManager) GenerateRefreshTokenWithExpiry(jti string, payload map[string]any, expiry time.Duration, subject ...string) (string, error) {
	return jtm.generateCustomToken(jti, payload, getSubject(subject, "refresh"), expiry)
}

// ValidateToken validates a JWT token
func (jtm *TokenManager) ValidateToken(tokenString string) (*jwtstd.Token, error) {
	if err := jtm.validateKey(); err != nil {
		return nil, err
	}

	return jwtstd.Parse(tokenString, func(token *jwtstd.Token) (any, error) {
		return []byte(jtm.key), nil
	})
}

// DecodeToken decodes a JWT token into its claims
func (jtm *TokenManager) DecodeToken(tokenString string) (map[string]any, error) {
	token, err := jtm.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return token.Claims.(jwtstd.MapClaims), nil
}

// GetTokenExpiryTime extracts the expiration time from a token
func (jtm *TokenManager) GetTokenExpiryTime(tokenString string) (time.Time, error) {
	claims, err := jtm.DecodeToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, ErrTokenParsing
	}

	return time.Unix(int64(exp), 0), nil
}

// IsTokenExpired checks if a token is expired
func (jtm *TokenManager) IsTokenExpired(tokenString string) (bool, error) {
	expiryTime, err := jtm.GetTokenExpiryTime(tokenString)
	if err != nil {
		return true, err
	}
	return expiryTime.Before(time.Now()), nil
}

// getPayloadFromClaims extracts the payload from token claims
func getPayloadFromClaims(claims map[string]any) (map[string]any, bool) {
	payloadAny, ok := claims["payload"]
	if !ok {
		return nil, false
	}
	payload, ok := payloadAny.(map[string]any)
	return payload, ok
}

// extractStringSlice extracts a string slice from the payload
func extractStringSlice(payload map[string]any, key string) []string {
	if valAny, ok := payload[key]; ok {
		if slice, ok := valAny.([]any); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// GetRolesFromToken extracts roles from token claims
func GetRolesFromToken(claims map[string]any) []string {
	if payload, ok := getPayloadFromClaims(claims); ok {
		return extractStringSlice(payload, "roles")
	}
	return []string{}
}

// GetPermissionsFromToken extracts permissions from token claims
func GetPermissionsFromToken(claims map[string]any) []string {
	if payload, ok := getPayloadFromClaims(claims); ok {
		return extractStringSlice(payload, "permissions")
	}
	return []string{}
}

// IsAdminFromToken checks if the token indicates an admin user
func IsAdminFromToken(claims map[string]any) bool {
	if payload, ok := getPayloadFromClaims(claims); ok {
		if isAdmin, ok := payload["is_admin"].(bool); ok {
			return isAdmin
		}
	}
	return false
}

// GetTenantIDFromToken gets the tenant ID from the token
func GetTenantIDFromToken(claims map[string]any) string {
	if payload, ok := getPayloadFromClaims(claims); ok {
		if tenantID, ok := payload["tenant_id"].(string); ok {
			return tenantID
		}
	}
	return ""
}

// GetUserIDFromToken gets the user ID from the token
func GetUserIDFromToken(claims map[string]any) string {
	if payload, ok := getPayloadFromClaims(claims); ok {
		if userID, ok := payload["user_id"].(string); ok {
			return userID
		}
	}
	return ""
}

// getSubject returns the subject if provided, otherwise returns the default subject
func getSubject(subject []string, defaultSubject string) string {
	if len(subject) > 0 {
		return subject[0]
	}
	return defaultSubject
}
