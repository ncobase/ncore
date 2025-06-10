package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MetricsConfig defines extension metrics configuration
type MetricsConfig struct {
	Enabled       bool           `json:"enabled" yaml:"enabled"`
	FlushInterval string         `json:"flush_interval" yaml:"flush_interval"`
	BatchSize     int            `json:"batch_size" yaml:"batch_size"`
	Retention     string         `json:"retention" yaml:"retention"`
	Storage       *StorageConfig `json:"storage" yaml:"storage"`
}

// StorageConfig defines metrics storage configuration
type StorageConfig struct {
	Type      string            `json:"type" yaml:"type"` // "memory", "redis", "auto"
	KeyPrefix string            `json:"key_prefix" yaml:"key_prefix"`
	Options   map[string]string `json:"options" yaml:"options"`
}

// GetDefaultMetricsConfig  provides default metrics configuration
func GetDefaultMetricsConfig(isDevelopment bool) *MetricsConfig {
	batchSize := 100
	retention := "7d"
	flushInterval := "30s"

	if isDevelopment {
		batchSize = 50
		retention = "24h"
		flushInterval = "10s"
	}

	return &MetricsConfig{
		Enabled:       false, // Default disabled
		FlushInterval: flushInterval,
		BatchSize:     batchSize,
		Retention:     retention,
		Storage: &StorageConfig{
			Type:      "auto",
			KeyPrefix: "ncore_ext",
			Options:   make(map[string]string),
		},
	}
}

// Validate validates the metrics configuration
func (m *MetricsConfig) Validate() error {
	if !m.Enabled {
		return nil
	}

	// Validate flush interval
	if m.FlushInterval != "" {
		if _, err := time.ParseDuration(m.FlushInterval); err != nil {
			return fmt.Errorf("invalid flush_interval: %v", err)
		}
	}

	// Validate retention
	if m.Retention != "" {
		if _, err := time.ParseDuration(m.Retention); err != nil {
			return fmt.Errorf("invalid retention: %v", err)
		}
	}

	// Validate batch size
	if m.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be greater than 0, got: %d", m.BatchSize)
	}

	// Validate storage config
	if m.Storage != nil {
		if err := m.Storage.Validate(); err != nil {
			return fmt.Errorf("storage config error: %v", err)
		}
	}

	return nil
}

// Validate validates the storage configuration
func (s *StorageConfig) Validate() error {
	validTypes := map[string]bool{
		"memory": true,
		"redis":  true,
		"auto":   true,
	}

	if !validTypes[s.Type] {
		return fmt.Errorf("invalid storage type: %s (valid: memory, redis, auto)", s.Type)
	}

	if s.KeyPrefix == "" {
		return fmt.Errorf("key_prefix cannot be empty")
	}

	return nil
}

// getMetricsConfig returns metrics configuration from viper
func getMetricsConfig(v *viper.Viper, isDevelopment bool) *MetricsConfig {
	defaultConfig := GetDefaultMetricsConfig(isDevelopment)

	if !v.IsSet("extension.metrics") {
		return defaultConfig
	}

	config := &MetricsConfig{
		Enabled:       getBoolWithDefault(v, "extension.metrics.enabled", defaultConfig.Enabled),
		FlushInterval: getStringWithDefault(v, "extension.metrics.flush_interval", defaultConfig.FlushInterval),
		BatchSize:     getIntWithDefault(v, "extension.metrics.batch_size", defaultConfig.BatchSize),
		Retention:     getStringWithDefault(v, "extension.metrics.retention", defaultConfig.Retention),
		Storage:       getStorageConfig(v, defaultConfig.Storage),
	}

	return config
}

// getStorageConfig returns storage configuration
func getStorageConfig(v *viper.Viper, defaultStorage *StorageConfig) *StorageConfig {
	if !v.IsSet("extension.metrics.storage") {
		return defaultStorage
	}

	config := &StorageConfig{
		Type:      getStringWithDefault(v, "extension.metrics.storage.type", defaultStorage.Type),
		KeyPrefix: getStringWithDefault(v, "extension.metrics.storage.key_prefix", defaultStorage.KeyPrefix),
		Options:   getStringMapWithDefault(v, "extension.metrics.storage.options", defaultStorage.Options),
	}

	return config
}

// getStringMapWithDefault returns string map value with default
func getStringMapWithDefault(v *viper.Viper, key string, defaultValue map[string]string) map[string]string {
	if v.IsSet(key) {
		return v.GetStringMapString(key)
	}
	return defaultValue
}
