//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/jwt"
)

// InitializeApp wires up the main application with all core dependencies.
// Returns the App, a cleanup function, and any initialization error.
//
// The cleanup function should be called when the application shuts down
// to properly release all resources (database connections, etc.).
//
// Usage:
//
//	app, cleanup, err := InitializeApp()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cleanup()
//	// Use app...
func InitializeApp() (*App, func(), error) {
	panic(wire.Build(
		// Configuration provider - loads and provides config
		config.ProviderSet,

		// Logger provider - provides *Logger with cleanup
		// Note: config.ProviderSet already provides *config.Logger which is
		// an alias for *logger/config.Config
		logger.ProviderSet,

		// Data layer provider - provides *Data with cleanup
		// Note: config.ProviderSet already provides *config.Data which is
		// an alias for *data/config.Config
		data.ProviderSet,

		// Extension manager provider - provides *Manager with cleanup
		manager.ProviderSet,

		// Application constructor
		NewApp,
	))
}

// ============================================================================
// JWT Token Manager Example
// ============================================================================

// InitializeTokenManager creates a JWT TokenManager with configuration.
// This is useful when you only need JWT functionality without the full app.
func InitializeTokenManager() (*jwt.TokenManager, error) {
	panic(wire.Build(
		config.ProviderSet,
		ProvideJWTConfig,
		jwt.ProvideTokenManager,
	))
}

// ============================================================================
// Worker Pool Example
// ============================================================================

// InitializeWorkerPool creates a worker pool with default configuration.
// Returns the pool, a cleanup function that gracefully stops workers, and any error.
func InitializeWorkerPool() (*worker.Pool, func(), error) {
	panic(wire.Build(
		ProvideDefaultWorkerConfig,
		worker.ProviderSet,
	))
}

// ============================================================================
// Custom Worker Pool Example
// ============================================================================

// InitializeCustomWorkerPool creates a worker pool with custom configuration.
func InitializeCustomWorkerPool(customCfg *CustomWorkerConfig) (*worker.Pool, func(), error) {
	panic(wire.Build(
		ProvideWorkerConfigFromCustom,
		worker.ProviderSet,
	))
}
