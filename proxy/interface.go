package proxy

import (
	"context"
	"net/http"
)

// Handler represents the handler for a proxy
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// RouteMatch represents a route matching configuration
type RouteMatch struct {
	Path    string   `json:"path"`
	Regex   string   `json:"regex,omitempty"`
	Methods []string `json:"methods,omitempty"`
}

// ServiceConfig represents the configuration for a proxied service
type ServiceConfig struct {
	Name       string            `json:"name"`
	TargetURL  string            `json:"targetUrl"`
	Routes     []RouteMatch      `json:"routes"`
	StripPath  string            `json:"stripPath,omitempty"`
	AddPath    string            `json:"addPath,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Middleware []string          `json:"middleware,omitempty"`
}

// ProxyConfig represents the configuration for a proxy
type ProxyConfig struct {
	Name     string          `json:"name"`
	Services []ServiceConfig `json:"services"`
}

// Interface defines the interface for a proxy
type Interface interface {
	Name() string
	Init(ctx context.Context, conf *ProxyConfig) error
	GetHandler() Handler
	GetConfig() *ProxyConfig
	MatchRoute(r *http.Request) bool
	ApplyMiddleware(next http.Handler) http.Handler
}
