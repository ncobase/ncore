package ctxutil

import (
	"context"
	"time"
)

const (
	// DefaultAsyncTimeout is the default timeout for async operations
	DefaultAsyncTimeout = 5 * time.Second
)

// WithAsyncContext creates a context suitable for async operations
// It derives from the parent context to preserve trace information,
// but uses a separate timeout to avoid blocking the main request
func WithAsyncContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultAsyncTimeout
	}

	// Create a detached context that won't be cancelled when parent is cancelled
	// but preserves trace information
	return context.WithTimeout(context.WithoutCancel(parent), timeout)
}

// WithAsyncContextDefault creates an async context with default timeout
func WithAsyncContextDefault(parent context.Context) (context.Context, context.CancelFunc) {
	return WithAsyncContext(parent, DefaultAsyncTimeout)
}
