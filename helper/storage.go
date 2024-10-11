package helper

import (
	"context"
	"ncobase/common/storage"

	"github.com/casdoor/oss"
)

// SetStorage sets storage to context.Context
func SetStorage(ctx context.Context, s oss.StorageInterface) context.Context {
	return SetValue(ctx, storageKey, s)
}

// GetStorage gets storage from context.Context
func GetStorage(ctx context.Context) (oss.StorageInterface, *storage.Config) {
	if s, ok := GetValue(ctx, storageKey).(oss.StorageInterface); ok {
		return s, GetConfig(ctx).Storage
	}

	// Get config
	storageConfig := GetConfig(ctx).Storage

	// Initialize storage
	s, err := storage.NewStorage(storageConfig)
	if err != nil {
		return nil, nil
	}

	// Set storage to context.Context
	ctx = SetStorage(ctx, s)
	return s, storageConfig
}
