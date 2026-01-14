package data

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/data/config"
)

// ProviderSet is the wire provider set for the data package.
// It provides *Data with a cleanup function that closes all connections.
//
// Usage:
//
//	wire.Build(
//	    data.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(ProvideData)

// ProvideData initializes and returns the data layer with cleanup function.
// The cleanup function should be called when the application shuts down
// to properly close all database connections and release resources.
func ProvideData(cfg *config.Config) (*Data, func(), error) {
	d, originalCleanup, err := New(cfg)
	if err != nil {
		return nil, nil, err
	}
	// Wrap the cleanup function to match Wire's expected signature
	cleanup := func() {
		originalCleanup()
	}
	return d, cleanup, nil
}
