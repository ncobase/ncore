package config

import (
	dc "ncobase/common/data/config"

	"github.com/spf13/viper"
)

// Data represents the data configuration
type Data = dc.Config

// DBNode represents a database node
type DBNode = dc.DBNode

// GetDataConfig returns data config
func getDataConfig(v *viper.Viper) *Data {
	return dc.GetConfig(v)
}
