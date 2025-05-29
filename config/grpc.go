package config

import "github.com/spf13/viper"

type GRPC struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`
}

func getGRPCConfig(v *viper.Viper) *GRPC {
	return &GRPC{
		Enabled: v.GetBool("grpc.enabled"),
		Host:    v.GetString("grpc.host"),
		Port:    v.GetInt("grpc.port"),
	}
}
