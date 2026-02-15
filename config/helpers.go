package config

import (
	"time"

	"github.com/spf13/viper"
)

// getDurationOrDefault returns duration from config or default value
func getDurationOrDefault(v *viper.Viper, key string, defaultValue time.Duration) time.Duration {
	if v.IsSet(key) {
		return v.GetDuration(key)
	}
	return defaultValue
}

// getUint32OrDefault returns uint32 from config or default value
func getUint32OrDefault(v *viper.Viper, key string, defaultValue uint32) uint32 {
	if v.IsSet(key) {
		return uint32(v.GetInt(key))
	}
	return defaultValue
}

// getIntOrDefault returns int from config or default value
func getIntOrDefault(v *viper.Viper, key string, defaultValue int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	return defaultValue
}

// getFloat64OrDefault returns float64 from config or default value
func getFloat64OrDefault(v *viper.Viper, key string, defaultValue float64) float64 {
	if v.IsSet(key) {
		return v.GetFloat64(key)
	}
	return defaultValue
}

// getStringOrDefault returns string from config or default value
func getStringOrDefault(v *viper.Viper, key string, defaultValue string) string {
	if v.IsSet(key) {
		return v.GetString(key)
	}
	return defaultValue
}

// getBoolOrDefault returns bool from config or default value
func getBoolOrDefault(v *viper.Viper, key string, defaultValue bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	return defaultValue
}
