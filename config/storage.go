package config

import (
	"ncobase/common/storage"

	"github.com/spf13/viper"
)

// Storage represents the storage configuration
type Storage = storage.Config

// getStorageConfig get storage config
func getStorageConfig(v *viper.Viper) *Storage {
	return storage.GetConfig(v)
}
