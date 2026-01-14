package manager

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/config"
)

// ProviderSet is the wire provider set for the extension manager package.
// It provides *Manager with a cleanup function that properly shuts down
// all extensions, services, and subsystems.
//
// Usage:
//
//	wire.Build(
//	    config.ProviderSet,
//	    manager.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(ProvideManager)

// ProvideManager initializes and returns the extension manager with cleanup function.
// The cleanup function should be called when the application shuts down
// to properly cleanup all extensions and release resources.
func ProvideManager(cfg *config.Config) (*Manager, func(), error) {
	m, err := NewManager(cfg)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		m.Cleanup()
	}
	return m, cleanup, nil
}
