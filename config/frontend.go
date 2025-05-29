package config

import "github.com/spf13/viper"

// Frontend frontend config struct
type Frontend struct {
	SignInURL string `json:"sign_in_url" yaml:"sign_in_url"`
	SignUpURL string `json:"sign_up_url" yaml:"sign_up_url"`
}

// FrontendConfig returns frontend config
func getFrontendConfig(v *viper.Viper) *Frontend {
	return &Frontend{
		SignInURL: v.GetString("frontend.sign_in_url"),
		SignUpURL: v.GetString("frontend.sign_up_url"),
	}
}
