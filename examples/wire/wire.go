//go:build wireinject
// +build wireinject

// Package main demonstrates how to use Google Wire with NCore's ProviderSets.
//
// This file contains the Wire injector definitions. Run `wire` command to generate
// the actual dependency injection code in wire_gen.go.
//
// To generate the wire_gen.go file:
//
//	go install github.com/google/wire/cmd/wire@latest
//	wire ./examples/wire/...
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

// App represents the main application with all core dependencies.
type App struct {
	Config  *config.Config
	Logger  *logger.Logger
	Data    *data.Data
	Manager *manager.Manager
}

// NewApp creates a new App instance with injected dependencies.
func NewApp(
	cfg *config.Config,
	log *logger.Logger,
	d *data.Data,
	m *manager.Manager,
) *App {
	return &App{
		Config:  cfg,
		Logger:  log,
		Data:    d,
		Manager: m,
	}
}

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

// ProvideJWTConfig extracts JWT configuration from the auth config.
func ProvideJWTConfig(auth *config.Auth) *jwt.Config {
	if auth == nil || auth.JWT == nil {
		return &jwt.Config{}
	}
	return &jwt.Config{
		Secret: auth.JWT.Secret,
	}
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

// ProvideDefaultWorkerConfig provides default worker pool configuration.
func ProvideDefaultWorkerConfig() *worker.Config {
	return worker.DefaultConfig()
}

// ============================================================================
// Custom Worker Pool Example
// ============================================================================

// CustomWorkerConfig represents custom worker configuration.
type CustomWorkerConfig struct {
	MaxWorkers int
	QueueSize  int
}

// InitializeCustomWorkerPool creates a worker pool with custom configuration.
func InitializeCustomWorkerPool(customCfg *CustomWorkerConfig) (*worker.Pool, func(), error) {
	panic(wire.Build(
		ProvideWorkerConfigFromCustom,
		worker.ProviderSet,
	))
}

// ProvideWorkerConfigFromCustom converts custom config to worker.Config.
func ProvideWorkerConfigFromCustom(custom *CustomWorkerConfig) *worker.Config {
	if custom == nil {
		return worker.DefaultConfig()
	}
	cfg := worker.DefaultConfig()
	if custom.MaxWorkers > 0 {
		cfg.MaxWorkers = custom.MaxWorkers
	}
	if custom.QueueSize > 0 {
		cfg.QueueSize = custom.QueueSize
	}
	return cfg
}
