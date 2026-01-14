package worker

import (
	"context"
	"time"

	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the worker package.
// It provides *Pool for task processing with proper lifecycle management.
//
// Usage:
//
//	wire.Build(
//	    worker.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(ProvidePool)

// ProvidePool creates a new worker Pool with cleanup function.
// The cleanup function gracefully stops the pool when called.
//
// If cfg is nil, default configuration is used:
//   - MaxWorkers: 10
//   - QueueSize: 1000
//   - TaskTimeout: 1 minute
func ProvidePool(cfg *Config) (*Pool, func(), error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	pool := NewPool(cfg)
	pool.Start()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		pool.Stop(ctx)
	}

	return pool, cleanup, nil
}

// ProvidePoolWithProcessor creates a worker Pool with a custom processor.
// This allows injecting custom task processing logic.
func ProvidePoolWithProcessor(cfg *Config, processor Processor) (*Pool, func(), error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	pool := NewPool(cfg, processor)
	pool.Start()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		pool.Stop(ctx)
	}

	return pool, cleanup, nil
}
