package config

import "github.com/spf13/viper"

// Auth auth config struct
type Auth struct {
	JWT       *JWT
	Casbin    *Casbin
	Whitelist []string
}

// getAuth returns the auth config.
func getAuth(v *viper.Viper) *Auth {
	return &Auth{
		JWT:       getJWT(v),
		Casbin:    getCasbin(v),
		Whitelist: getWhitelist(v),
	}
}

// JWT jwt config struct
type JWT struct {
	Secret string
	Expire int
}

// getJWT returns the jwt config.
func getJWT(v *viper.Viper) *JWT {
	return &JWT{
		Secret: v.GetString("auth.jwt.secret"),
		Expire: v.GetInt("auth.jwt.expire"),
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
