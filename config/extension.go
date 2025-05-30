package config

import (
	ec "github.com/ncobase/ncore/extension/config"
	"github.com/spf13/viper"
)

// Extension represents the extension configuration
type Extension = ec.Config

// getExtensionConfig returns the extension config
func getExtensionConfig(v *viper.Viper) *Extension {
	return ec.GetConfig(v)
}
