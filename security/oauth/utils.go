package oauth

import (
	"fmt"
	"reflect"
	"strings"
)

// getString safely extracts string value from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		// Try to convert to string
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// getBool safely extracts boolean value from map
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
		// Try to convert common values
		switch fmt.Sprintf("%v", v) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no", "":
			return false
		}
	}
	return false
}

// getInt safely extracts integer value from map
func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case string:
			// Could add string parsing if needed
		}
	}
	return 0
}

// getStringSlice safely extracts string slice from map
func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		if slice, ok := v.([]string); ok {
			return slice
		}
	}
	return nil
}

// MergeParams merges multiple parameter maps
func MergeParams(params ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, p := range params {
		for k, v := range p {
			result[k] = v
		}
	}
	return result
}

// ValidateRequiredScopes checks if required scopes are present
func ValidateRequiredScopes(provider string, requestedScopes []string) error {
	info := GetProviderInfo(provider)
	if len(info.RequiredScopes) == 0 {
		return nil
	}

	scopeMap := make(map[string]bool)
	for _, scope := range requestedScopes {
		scopeMap[scope] = true
	}

	for _, required := range info.RequiredScopes {
		if !scopeMap[required] {
			return fmt.Errorf("required scope '%s' missing for provider %s", required, provider)
		}
	}

	return nil
}

// NormalizeProviderName normalizes provider name to lowercase
func NormalizeProviderName(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// IsProviderEnabled checks if provider is enabled in config
func IsProviderEnabled(config *Config, provider string) bool {
	if config == nil || config.Providers == nil {
		return false
	}

	providerConfig, exists := config.Providers[provider]
	return exists && providerConfig.Enabled && providerConfig.ClientID != ""
}

// GetEnabledProviders returns list of enabled providers
func GetEnabledProviders(config *Config) []string {
	if config == nil || config.Providers == nil {
		return nil
	}

	var enabled []string
	for name, providerConfig := range config.Providers {
		if providerConfig.Enabled && providerConfig.ClientID != "" {
			enabled = append(enabled, name)
		}
	}

	return enabled
}

// ValidateConfig validates OAuth configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("OAuth config is nil")
	}

	if config.Providers == nil {
		return fmt.Errorf("no OAuth providers configured")
	}

	for name, providerConfig := range config.Providers {
		if err := validateProviderConfig(name, providerConfig); err != nil {
			return fmt.Errorf("provider %s: %w", name, err)
		}
	}

	return nil
}

// validateProviderConfig validates individual provider configuration
func validateProviderConfig(name string, config *ProviderConfig) error {
	if !config.Enabled {
		return nil // Skip validation for disabled providers
	}

	if config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if config.ClientSecret == "" && name != "apple" { // Apple uses JWT
		return fmt.Errorf("client_secret is required")
	}

	if config.RedirectURL == "" {
		return fmt.Errorf("redirect_url is required")
	}

	if config.AuthURL == "" {
		return fmt.Errorf("auth_url is required")
	}

	if config.TokenURL == "" {
		return fmt.Errorf("token_url is required")
	}

	// Validate scopes if provider has required scopes
	if err := ValidateRequiredScopes(name, config.Scopes); err != nil {
		return err
	}

	return nil
}

// CloneConfig creates a deep copy of OAuth config
func CloneConfig(config *Config) *Config {
	if config == nil {
		return nil
	}

	cloned := &Config{
		Providers:    make(map[string]*ProviderConfig),
		DefaultScope: make([]string, len(config.DefaultScope)),
		EnablePKCE:   config.EnablePKCE,
		StateSecret:  config.StateSecret,
	}

	copy(cloned.DefaultScope, config.DefaultScope)

	for name, providerConfig := range config.Providers {
		cloned.Providers[name] = cloneProviderConfig(providerConfig)
	}

	return cloned
}

// cloneProviderConfig creates a deep copy of provider config
func cloneProviderConfig(config *ProviderConfig) *ProviderConfig {
	if config == nil {
		return nil
	}

	cloned := &ProviderConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       make([]string, len(config.Scopes)),
		AuthURL:      config.AuthURL,
		TokenURL:     config.TokenURL,
		UserInfoURL:  config.UserInfoURL,
		RevokeURL:    config.RevokeURL,
		Enabled:      config.Enabled,
		ExtraParams:  make(map[string]string),
	}

	copy(cloned.Scopes, config.Scopes)

	for k, v := range config.ExtraParams {
		cloned.ExtraParams[k] = v
	}

	return cloned
}

// CompareConfigs compares two OAuth configurations
func CompareConfigs(a, b *Config) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a, b)
}
