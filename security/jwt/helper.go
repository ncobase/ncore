package jwt

import "time"

// GetString safely extracts string value from any map
func GetString(data map[string]any, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// GetBool safely extracts boolean value from any map
func GetBool(data map[string]any, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

// GetInt64 safely extracts int64 value from any map
func GetInt64(data map[string]any, key string) int64 {
	switch val := data[key].(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int32:
		return int64(val)
	}
	return 0
}

// GetInt safely extracts int value from any map
func GetInt(data map[string]any, key string) int {
	return int(GetInt64(data, key))
}

// GetFloat64 safely extracts float64 value from any map
func GetFloat64(data map[string]any, key string) float64 {
	switch val := data[key].(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int:
		return float64(val)
	}
	return 0
}

// GetStringSlice safely extracts string slice from any map
func GetStringSlice(data map[string]any, key string) []string {
	if val, ok := data[key].([]any); ok {
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

// GetMap safely extracts nested map from any map
func GetMap(data map[string]any, key string) map[string]any {
	if val, ok := data[key].(map[string]any); ok {
		return val
	}
	return map[string]any{}
}

// GetTokenID extracts JWT ID (jti) from token claims
func GetTokenID(claims map[string]any) string {
	return GetString(claims, "jti")
}

// GetSubject extracts subject (sub) from token claims
func GetSubject(claims map[string]any) string {
	return GetString(claims, "sub")
}

// GetIssuer extracts issuer (iss) from token claims
func GetIssuer(claims map[string]any) string {
	return GetString(claims, "iss")
}

// GetAudience extracts audience (aud) from token claims
func GetAudience(claims map[string]any) []string {
	return GetStringSlice(claims, "aud")
}

// GetExpiration extracts expiration time from token claims
func GetExpiration(claims map[string]any) time.Time {
	if exp := GetFloat64(claims, "exp"); exp > 0 {
		return time.Unix(int64(exp), 0)
	}
	return time.Time{}
}

// GetIssuedAt extracts issued at time from token claims
func GetIssuedAt(claims map[string]any) time.Time {
	if iat := GetFloat64(claims, "iat"); iat > 0 {
		return time.Unix(int64(iat), 0)
	}
	return time.Time{}
}

// GetNotBefore extracts not before time from token claims
func GetNotBefore(claims map[string]any) time.Time {
	if nbf := GetFloat64(claims, "nbf"); nbf > 0 {
		return time.Unix(int64(nbf), 0)
	}
	return time.Time{}
}

// GetPayload extracts payload from token claims
func GetPayload(claims map[string]any) map[string]any {
	return GetMap(claims, "payload")
}

// IsAccessToken checks if token is an access token
func IsAccessToken(claims map[string]any) bool {
	return GetSubject(claims) == "access"
}

// IsRefreshToken checks if token is a refresh token
func IsRefreshToken(claims map[string]any) bool {
	return GetSubject(claims) == "refresh"
}

// IsTokenExpired checks if token is expired based on claims
func IsTokenExpired(claims map[string]any) bool {
	exp := GetExpiration(claims)
	return !exp.IsZero() && exp.Before(time.Now())
}

// IsTokenActive checks if token is currently active (not before current time)
func IsTokenActive(claims map[string]any) bool {
	nbf := GetNotBefore(claims)
	return nbf.IsZero() || nbf.Before(time.Now())
}

// IsTokenStale checks if token is older than specified duration
func IsTokenStale(claims map[string]any, staleDuration time.Duration) bool {
	issuedAt := GetIssuedAt(claims)
	if issuedAt.IsZero() {
		return true
	}
	return time.Since(issuedAt) > staleDuration
}

// ValidateTokenType ensures token is of expected type
func ValidateTokenType(claims map[string]any, expectedType string) error {
	actualType := GetSubject(claims)
	if actualType != expectedType {
		return TokenError("invalid token type: expected " + expectedType + ", got " + actualType)
	}
	return nil
}

// ValidateTokenTiming validates token timing (exp, iat, nbf)
func ValidateTokenTiming(claims map[string]any) error {
	now := time.Now()

	// Check expiration
	if exp := GetExpiration(claims); !exp.IsZero() && exp.Before(now) {
		return ErrTokenExpired
	}

	// Check not before
	if nbf := GetNotBefore(claims); !nbf.IsZero() && nbf.After(now) {
		return TokenError("token not yet valid")
	}

	return nil
}

// GetPayloadString extracts string value from payload
func GetPayloadString(claims map[string]any, key string) string {
	payload := GetPayload(claims)
	return GetString(payload, key)
}

// GetPayloadBool extracts boolean value from payload
func GetPayloadBool(claims map[string]any, key string) bool {
	payload := GetPayload(claims)
	return GetBool(payload, key)
}

// GetPayloadInt extracts int value from payload
func GetPayloadInt(claims map[string]any, key string) int {
	payload := GetPayload(claims)
	return GetInt(payload, key)
}

// GetPayloadStringSlice extracts string slice from payload
func GetPayloadStringSlice(claims map[string]any, key string) []string {
	payload := GetPayload(claims)
	return GetStringSlice(payload, key)
}

// HasPayloadValue checks if payload contains a specific key with non-empty value
func HasPayloadValue(claims map[string]any, key string) bool {
	payload := GetPayload(claims)
	_, exists := payload[key]
	return exists
}

// ContainsValue checks if a slice contains a specific value
func ContainsValue(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsAnyValue checks if a slice contains any of the specified values
func ContainsAnyValue(slice []string, values ...string) bool {
	for _, item := range slice {
		for _, value := range values {
			if item == value {
				return true
			}
		}
	}
	return false
}
