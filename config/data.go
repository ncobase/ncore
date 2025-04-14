package config

import (
	config3 "github.com/ncobase/ncore/data/config"
	"github.com/spf13/viper"
)

// Data represents the data configuration
type Data = config3.Config

// DBNode represents a database node
type DBNode = config3.DBNode

// GetDataConfig returns data config
func getDataConfig(v *viper.Viper) *Data {
	return config3.GetConfig(v)
}
