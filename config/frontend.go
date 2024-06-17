package config

import "github.com/spf13/viper"

// Frontend frontend config struct
type Frontend struct {
	SignInURL string
	SignUpURL string
}

// FrontendConfig returns frontend config
func getFrontendConfig(v *viper.Viper) Frontend {
	return Frontend{
		SignInURL: v.GetString("frontend.sign_in_url"),
		SignUpURL: v.GetString("frontend.sign_up_url"),
	}
}
