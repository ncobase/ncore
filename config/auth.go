package config

import (
	"time"

	"github.com/spf13/viper"
)

// Auth auth config struct
type Auth struct {
	JWT                    *JWT     `json:"jwt" yaml:"jwt"`
	Casbin                 *Casbin  `json:"casbin" yaml:"casbin"`
	Whitelist              []string `json:"whitelist" yaml:"whitelist"`
	MaxSessions            int      `json:"max_sessions" yaml:"max_sessions"`
	SessionCleanupInterval int      `json:"session_cleanup_interval" yaml:"session_cleanup_interval"`
}

// getAuth returns the auth config.
func getAuth(v *viper.Viper) *Auth {
	return &Auth{
		JWT:                    getJWT(v),
		Casbin:                 getCasbin(v),
		Whitelist:              getWhitelist(v),
		MaxSessions:            v.GetInt("auth.max_sessions"),
		SessionCleanupInterval: v.GetInt("auth.session_cleanup_interval"),
	}
}

// JWT jwt config struct
type JWT struct {
	Secret string
	Expiry time.Duration
}

// getJWT returns the jwt config.
func getJWT(v *viper.Viper) *JWT {
	return &JWT{
		Secret: v.GetString("auth.jwt.secret"),
		Expiry: v.GetDuration("auth.jwt.expiry"),
	}
}

// Casbin casbin config struct
type Casbin struct {
	Path  string
	Model string
}

// getCasbin returns the casbin config.
func getCasbin(v *viper.Viper) *Casbin {
	return &Casbin{
		Path:  v.GetString("auth.casbin.path"),
		Model: v.GetString("auth.casbin.model"),
	}
}

// getWhitelist returns the whitelist config.
func getWhitelist(v *viper.Viper) []string {
	return v.GetStringSlice("auth.whitelist")
}
