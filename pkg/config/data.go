package config

import (
	config2 "github.com/ncobase/ncore/pkg/data/config"

	"github.com/spf13/viper"
)

// Data represents the data configuration
type Data = config2.Config

// DBNode represents a database node
type DBNode = config2.DBNode

// GetDataConfig returns data config
func getDataConfig(v *viper.Viper) *Data {
	return config2.GetConfig(v)
}
