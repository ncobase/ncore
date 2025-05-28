package jwt

import "time"

// getPayload extracts payload from token claims
func getPayload(claims map[string]any) (map[string]any, bool) {
	if payload, ok := claims["payload"].(map[string]any); ok {
		return payload, true
	}
	return nil, false
}

// getString safely extracts string value from payload
func getString(payload map[string]any, key string) string {
	if val, ok := payload[key].(string); ok {
		return val
	}
	return ""
}

// getBool safely extracts boolean value from payload
func getBool(payload map[string]any, key string) bool {
	if val, ok := payload[key].(bool); ok {
		return val
	}
	return false
}

// getInt64 safely extracts int64 value from payload
func getInt64(payload map[string]any, key string) int64 {
	switch val := payload[key].(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	}
	return 0
}

// getInt safely extracts int value from payload
func getInt(payload map[string]any, key string) int {
	if val, ok := payload[key].(int); ok {
		return val
	}
	return 0
}

// getStringSlice safely extracts string slice from payload
func getStringSlice(payload map[string]any, key string) []string {
	if val, ok := payload[key].([]any); ok {
		result := make([]string, 0, len(val))
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// GetTokenIDFromToken extracts JWT ID (jti) from token claims
func GetTokenIDFromToken(claims map[string]any) string {
	if jti, ok := claims["jti"].(string); ok {
		return jti
	}
	return ""
}

// GetSubjectFromToken extracts subject (sub) from token claims
func GetSubjectFromToken(claims map[string]any) string {
	if sub, ok := claims["sub"].(string); ok {
		return sub
	}
	return ""
}

// GetExpirationFromToken extracts expiration time from token claims
func GetExpirationFromToken(claims map[string]any) time.Time {
	if exp, ok := claims["exp"].(float64); ok && exp > 0 {
		return time.Unix(int64(exp), 0)
	}
	return time.Time{}
}

// GetUserIDFromToken extracts user ID from token claims
func GetUserIDFromToken(claims map[string]any) string {
	if payload, ok := getPayload(claims); ok {
		return getString(payload, "user_id")
	}
	return ""
}

// GetUsernameFromToken extracts username from token claims
func GetUsernameFromToken(claims map[string]any) string {
	if payload, ok := getPayload(claims); ok {
		return getString(payload, "username")
	}
	return ""
}

// GetEmailFromToken extracts email from token claims
func GetEmailFromToken(claims map[string]any) string {
	if payload, ok := getPayload(claims); ok {
		return getString(payload, "email")
	}
	return ""
}

// GetTenantIDFromToken extracts tenant ID from token claims
func GetTenantIDFromToken(claims map[string]any) string {
	if payload, ok := getPayload(claims); ok {
		return getString(payload, "tenant_id")
	}
	return ""
}

// GetTenantIDsFromToken extracts tenant IDs from token claims
func GetTenantIDsFromToken(claims map[string]any) []string {
	if payload, ok := getPayload(claims); ok {
		return getStringSlice(payload, "tenant_ids")
	}
	return []string{}
}

// GetRolesFromToken extracts roles from token claims
func GetRolesFromToken(claims map[string]any) []string {
	if payload, ok := getPayload(claims); ok {
		return getStringSlice(payload, "roles")
	}
	return []string{}
}

// GetPermissionsFromToken extracts permissions from token claims
func GetPermissionsFromToken(claims map[string]any) []string {
	if payload, ok := getPayload(claims); ok {
		return getStringSlice(payload, "permissions")
	}
	return []string{}
}

// IsAdminFromToken checks if user is admin from token claims
func IsAdminFromToken(claims map[string]any) bool {
	if payload, ok := getPayload(claims); ok {
		return getBool(payload, "is_admin")
	}
	return false
}

// GetUserStatusFromToken extracts user status from token claims
func GetUserStatusFromToken(claims map[string]any) int {
	if payload, ok := getPayload(claims); ok {
		return getInt(payload, "user_status")
	}
	return 0
}

// IsCertifiedFromToken checks if user is certified from token claims
func IsCertifiedFromToken(claims map[string]any) bool {
	if payload, ok := getPayload(claims); ok {
		return getBool(payload, "is_certified")
	}
	return false
}

// GetIssuedAtFromToken extracts issued at time from token claims
func GetIssuedAtFromToken(claims map[string]any) time.Time {
	if iat, ok := claims["iat"].(float64); ok && iat > 0 {
		return time.Unix(int64(iat), 0)
	}
	return time.Time{}
}

// ValidateTokenUser validates user info in token against current user data
func ValidateTokenUser(claims map[string]any, currentUser *TokenUser) error {
	payload, ok := getPayload(claims)
	if !ok {
		return ErrInvalidToken
	}

	// Validate user ID
	if tokenUserID := getString(payload, "user_id"); tokenUserID != currentUser.ID {
		return TokenError("user ID mismatch")
	}

	// Validate username
	if tokenUsername := getString(payload, "username"); tokenUsername != currentUser.Username {
		return TokenError("username mismatch")
	}

	// Validate user status
	if tokenStatus := getInt(payload, "user_status"); tokenStatus != currentUser.Status {
		return TokenError("user status changed")
	}

	return nil
}

// TokenUser represents minimal user info for validation
type TokenUser struct {
	ID       string
	Username string
	Email    string
	Status   int
}

// IsTokenStale checks if token is older than specified duration
func IsTokenStale(claims map[string]any, staleDuration time.Duration) bool {
	issuedAt := GetIssuedAtFromToken(claims)
	if issuedAt.IsZero() {
		return true
	}
	return time.Since(issuedAt) > staleDuration
}

// HasRole checks if user has specific role in token
func HasRole(claims map[string]any, role string) bool {
	roles := GetRolesFromToken(claims)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if user has specific permission in token
func HasPermission(claims map[string]any, permission string) bool {
	permissions := GetPermissionsFromToken(claims)
	for _, p := range permissions {
		if p == permission || p == "*:*" {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(claims map[string]any, roles ...string) bool {
	userRoles := GetRolesFromToken(claims)
	for _, userRole := range userRoles {
		for _, role := range roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

// IsAdminRole checks if user has admin role
func IsAdminRole(claims map[string]any) bool {
	return HasAnyRole(claims, "super-admin", "system-admin")
}

// IsAccessToken checks if token is an access token
func IsAccessToken(claims map[string]any) bool {
	return GetSubjectFromToken(claims) == "access"
}

// IsRefreshToken checks if token is a refresh token
func IsRefreshToken(claims map[string]any) bool {
	return GetSubjectFromToken(claims) == "refresh"
}

// IsRegisterToken checks if token is a register token
func IsRegisterToken(claims map[string]any) bool {
	subject := GetSubjectFromToken(claims)
	return subject != "access" && subject != "refresh"
}

// ValidateTokenType ensures token is of expected type
func ValidateTokenType(claims map[string]any, expectedType string) error {
	actualType := GetSubjectFromToken(claims)
	if actualType != expectedType {
		return TokenError("invalid token type: expected " + expectedType + ", got " + actualType)
	}
	return nil
}
