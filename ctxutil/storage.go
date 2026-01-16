package ctxutil

import (
	"context"

	"github.com/ncobase/ncore/oss"
)

func SetStorage(ctx context.Context, s oss.Interface) context.Context {
	return SetValue(ctx, storageKey, s)
}

func GetStorage(ctx context.Context) (oss.Interface, *oss.Config) {
	if s, ok := GetValue(ctx, storageKey).(oss.Interface); ok {
		return s, GetConfig(ctx).Storage
	}

	storageConfig := GetConfig(ctx).Storage

	s, err := oss.NewStorage(storageConfig)
	if err != nil {
		return nil, nil
	}

	ctx = SetStorage(ctx, s)
	return s, storageConfig
}
