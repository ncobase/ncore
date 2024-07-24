package proxy

import (
	"net/http"
	"sync"
)

// RegistryFunc is a function type for middleware
type RegistryFunc func(http.Handler) http.Handler

// RegistryInterface is an interface for registering and retrieving middleware
type RegistryInterface interface {
	Register(name string, middleware RegistryFunc)
	Get(name string) (RegistryFunc, bool)
}

// Registry is a basic implementation of the RegistryInterface interface
type Registry struct {
	middleware map[string]RegistryFunc
	mu         sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		middleware: make(map[string]RegistryFunc),
	}
}

// Register adds a new middleware to the registry
func (r *Registry) Register(name string, middleware RegistryFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware[name] = middleware
}

// Get retrieves a middleware from the registry
func (r *Registry) Get(name string) (RegistryFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mw, ok := r.middleware[name]
	return mw, ok
}
