package config

import (
	"time"

	"github.com/spf13/viper"
)

// Messaging messaging config for all message channels
type Messaging struct {
	PublishTimeout   time.Duration `json:"publish_timeout" yaml:"publish_timeout"`
	CrossRegionMode  bool          `json:"cross_region_mode" yaml:"cross_region_mode"`
	RetryAttempts    int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryBackoffMax  time.Duration `json:"retry_backoff_max" yaml:"retry_backoff_max"`
	FallbackToMemory bool          `json:"fallback_to_memory" yaml:"fallback_to_memory"`
}

// getMessagingConfig reads messaging configurations
func getMessagingConfig(v *viper.Viper) *Messaging {
	return &Messaging{
		PublishTimeout:   getMessagingPublishTimeout(v),
		CrossRegionMode:  getMessagingCrossRegionMode(v),
		RetryAttempts:    getMessagingRetryAttempts(v),
		RetryBackoffMax:  getMessagingRetryBackoffMax(v),
		FallbackToMemory: getMessagingFallbackToMemory(v),
	}
}

// getMessagingPublishTimeout gets publish timeout with defaults
func getMessagingPublishTimeout(v *viper.Viper) time.Duration {
	if v.IsSet("data.messaging.publish_timeout") {
		return v.GetDuration("data.messaging.publish_timeout")
	}

	// Default based on cross-region mode
	if getMessagingCrossRegionMode(v) {
		return 60 * time.Second
	}

	return 30 * time.Second
}

// getMessagingCrossRegionMode gets cross-region mode with default
func getMessagingCrossRegionMode(v *viper.Viper) bool {
	if v.IsSet("data.messaging.cross_region_mode") {
		return v.GetBool("data.messaging.cross_region_mode")
	}
	return false
}

// getMessagingRetryAttempts gets retry attempts with default
func getMessagingRetryAttempts(v *viper.Viper) int {
	if v.IsSet("data.messaging.retry_attempts") {
		return v.GetInt("data.messaging.retry_attempts")
	}
	return 3
}

// getMessagingRetryBackoffMax gets max retry backoff with default
func getMessagingRetryBackoffMax(v *viper.Viper) time.Duration {
	if v.IsSet("data.messaging.retry_backoff_max") {
		return v.GetDuration("data.messaging.retry_backoff_max")
	}
	return 30 * time.Second
}

// getMessagingFallbackToMemory gets fallback setting with default
func getMessagingFallbackToMemory(v *viper.Viper) bool {
	if v.IsSet("data.messaging.fallback_to_memory") {
		return v.GetBool("data.messaging.fallback_to_memory")
	}
	return true
}
