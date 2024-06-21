package config

import (
	"ncobase/common/storage"

	"github.com/spf13/viper"
)

// getStorageConfig get storage config
func getStorageConfig(v *viper.Viper) storage.Config {
	return storage.Config{
		Provider: v.GetString("storage.provider"),
		ID:       v.GetString("storage.id"),
		Secret:   v.GetString("storage.secret"),
		Region:   v.GetString("storage.region"),
		Bucket:   v.GetString("storage.bucket"),
		Endpoint: v.GetString("storage.endpoint"),
	}
}
