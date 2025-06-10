package config

import "github.com/spf13/viper"

// Metrics data metrics config
type Metrics struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	StorageType   string `yaml:"storage_type" json:"storage_type"` // "memory", "redis"
	KeyPrefix     string `yaml:"key_prefix" json:"key_prefix"`
	RetentionDays int    `yaml:"retention_days" json:"retention_days"`
	BatchSize     int    `yaml:"batch_size" json:"batch_size"`
}

// getMetricsConfig returns metrics config
func getMetricsConfig(v *viper.Viper) *Metrics {
	enabled := v.GetBool("data.metrics.enabled")
	storageType := v.GetString("data.metrics.storage_type")
	if storageType == "" {
		if enabled && v.IsSet("data.redis.addr") && v.GetString("data.redis.addr") != "" {
			storageType = "redis"
		} else {
			storageType = "memory"
		}
	}

	return &Metrics{
		Enabled:       enabled,
		StorageType:   storageType,
		KeyPrefix:     getStringOrDefault(v, "data.metrics.key_prefix", "ncore_data"),
		RetentionDays: getIntOrDefault(v, "data.metrics.retention_days", 7),
		BatchSize:     getIntOrDefault(v, "data.metrics.batch_size", 100),
	}
}

// getStringOrDefault returns string value or default
func getStringOrDefault(v *viper.Viper, key, defaultValue string) string {
	if v.IsSet(key) {
		return v.GetString(key)
	}
	return defaultValue
}

// getIntOrDefault returns int value or default
func getIntOrDefault(v *viper.Viper, key string, defaultValue int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	return defaultValue
}
