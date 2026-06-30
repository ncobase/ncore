package storage

import (
	"context"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/oss"
)

const storageKey = "storage"

// Set stores an object storage client in context.
func Set(ctx context.Context, s oss.Interface) context.Context {
	return ctxutil.SetValue(ctx, storageKey, s)
}

// Get returns the configured object storage client and its source config.
func Get(ctx context.Context) (oss.Interface, *config.Storage) {
	if s, ok := ctxutil.GetValue(ctx, storageKey).(oss.Interface); ok {
		cfg := ctxutil.GetConfig(ctx)
		if cfg == nil {
			return s, nil
		}
		return s, cfg.Storage
	}

	cfg := ctxutil.GetConfig(ctx)
	if cfg == nil || cfg.Storage == nil {
		return nil, nil
	}
	storageConfig := cfg.Storage
	s, err := oss.NewStorage(ToOSSConfig(storageConfig))
	if err != nil {
		return nil, nil
	}

	ctx = Set(ctx, s)
	return s, storageConfig
}

// ToOSSConfig converts lightweight application config to an oss runtime config.
func ToOSSConfig(storageConfig *config.Storage) *oss.Config {
	if storageConfig == nil {
		return nil
	}
	return &oss.Config{
		Provider:           storageConfig.Provider,
		ID:                 storageConfig.ID,
		Secret:             storageConfig.Secret,
		Region:             storageConfig.Region,
		Bucket:             storageConfig.Bucket,
		Endpoint:           storageConfig.Endpoint,
		ServiceAccountJSON: storageConfig.ServiceAccountJSON,
		SharedFolder:       storageConfig.SharedFolder,
		OtpCode:            storageConfig.OtpCode,
		Debug:              storageConfig.Debug,
		AppID:              storageConfig.AppID,
	}
}
