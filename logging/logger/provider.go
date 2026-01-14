package logger

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/logging/logger/config"
)

// ProviderSet is the wire provider set for the logger package
var ProviderSet = wire.NewSet(ProvideLogger)

// ProvideLogger initializes and returns the standard logger
func ProvideLogger(cfg *config.Config) (*Logger, func(), error) {
	cleanup, err := New(cfg)
	return StdLogger(), cleanup, err
}
