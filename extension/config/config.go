package config

import (
	nconfig "github.com/ncobase/ncore/config"
	"github.com/spf13/viper"
)

// BuiltInMode represents the build tag value for built-in extension mode.
const BuiltInMode = nconfig.ExtensionBuiltInMode

// Config contains extension system infrastructure settings.
type Config = nconfig.Extension

// SecurityConfig contains extension sandbox and trust settings.
type SecurityConfig = nconfig.ExtensionSecurity

// PerformanceConfig contains extension runtime limits.
type PerformanceConfig = nconfig.ExtensionPerformance

// MetricsConfig contains extension metrics collection settings.
type MetricsConfig = nconfig.ExtensionMetrics

// StorageConfig contains extension metrics persistence settings.
type StorageConfig = nconfig.ExtensionMetricsStorage

// GetConfig returns extension configuration and panics on invalid values.
func GetConfig(v *viper.Viper) *Config {
	return nconfig.GetExtensionConfig(v)
}

// GetConfigE returns extension configuration and validation errors.
func GetConfigE(v *viper.Viper) (*Config, error) {
	return nconfig.GetExtensionConfigE(v)
}
