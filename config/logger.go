package config

import (
	lc "github.com/ncobase/ncore/logging/logger/config"
	"github.com/spf13/viper"
)

// Logger represents the logger configuration
type Logger = lc.Config

// getLoggerConfig returns the logger configuration
func getLoggerConfig(v *viper.Viper) *Logger {
	return lc.GetConfig(v)
}
