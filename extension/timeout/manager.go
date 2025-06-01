package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/extension/config"
)

// Manager handles timeout operations for extension loading and initialization
type Manager struct {
	loadTimeout       time.Duration
	initTimeout       time.Duration
	dependencyTimeout time.Duration
}

// NewManager creates a new timeout manager
func NewManager(cfg *config.Config) (*Manager, error) {
	loadTimeout := 30 * time.Second // default
	if cfg.LoadTimeout != "" {
		if parsed, err := time.ParseDuration(cfg.LoadTimeout); err != nil {
			return nil, fmt.Errorf("invalid load timeout: %v", err)
		} else {
			loadTimeout = parsed
		}
	}

	initTimeout := 60 * time.Second // default
	if cfg.InitTimeout != "" {
		if parsed, err := time.ParseDuration(cfg.InitTimeout); err != nil {
			return nil, fmt.Errorf("invalid init timeout: %v", err)
		} else {
			initTimeout = parsed
		}
	}

	dependencyTimeout := 15 * time.Second // default
	if cfg.DependencyTimeout != "" {
		if parsed, err := time.ParseDuration(cfg.DependencyTimeout); err != nil {
			return nil, fmt.Errorf("invalid dependency timeout: %v", err)
		} else {
			dependencyTimeout = parsed
		}
	}

	return &Manager{
		loadTimeout:       loadTimeout,
		initTimeout:       initTimeout,
		dependencyTimeout: dependencyTimeout,
	}, nil
}

// WithLoadTimeout executes function with load timeout
func (tm *Manager) WithLoadTimeout(ctx context.Context, fn func(context.Context) error) error {
	return tm.withTimeout(ctx, tm.loadTimeout, "load", fn)
}

// WithInitTimeout executes function with initialization timeout
func (tm *Manager) WithInitTimeout(ctx context.Context, fn func(context.Context) error) error {
	return tm.withTimeout(ctx, tm.initTimeout, "initialization", fn)
}

// WithDependencyTimeout executes function with dependency resolution timeout
func (tm *Manager) WithDependencyTimeout(ctx context.Context, fn func(context.Context) error) error {
	return tm.withTimeout(ctx, tm.dependencyTimeout, "dependency resolution", fn)
}

// withTimeout executes function with specified timeout
func (tm *Manager) withTimeout(ctx context.Context, timeout time.Duration, operation string, fn func(context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(timeoutCtx)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("%s timeout after %v", operation, timeout)
	}
}
