package config

import (
	oc "github.com/ncobase/ncore/security/oauth"
	"github.com/spf13/viper"
)

// OAuth represents the OAuth configuration
type OAuth = oc.Config

// getOAuthConfig returns the OAuth configuration
func getOAuthConfig(v *viper.Viper) *OAuth {
	return oc.GetConfig(v)
}
