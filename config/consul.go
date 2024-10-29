package config

import "github.com/spf13/viper"

// Consul config struct
type Consul struct {
	Address string
	Scheme  string
}

// getConsulConfig get consul config
func getConsulConfig(v *viper.Viper) *Consul {
	return &Consul{
		Address: v.GetString("consul.address"),
		Scheme:  v.GetString("consul.scheme"),
	}
}
