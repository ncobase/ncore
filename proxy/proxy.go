package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

// Proxy represents a HTTP reverse proxy
type Proxy struct {
	config     *ProxyConfig
	handler    http.Handler
	routes     []*regexp.Regexp
	eventBus   EventBusInterface
	middleware []RegistryFunc
	registry   RegistryInterface
}

// NewProxy creates a new HTTP proxy
func NewProxy(ctx context.Context, conf *ProxyConfig, eventBus EventBusInterface, registry RegistryInterface) (*Proxy, error) {
	proxy := &Proxy{
		config:   conf,
		eventBus: eventBus,
		registry: registry,
	}

	director := func(req *http.Request) {
		for _, service := range conf.Services {
			for _, route := range service.Routes {
				if matchRoute(req, route) {
					targetURL, _ := url.Parse(service.TargetURL)
					req.URL.Scheme = targetURL.Scheme
					req.URL.Host = targetURL.Host

					if service.StripPath != "" {
						req.URL.Path = strings.TrimPrefix(req.URL.Path, service.StripPath)
					}
					if service.AddPath != "" {
						req.URL.Path = service.AddPath + req.URL.Path
					}

					// Add custom headers
					for key, value := range service.Headers {
						req.Header.Set(key, value)
					}

					proxy.eventBus.Publish(ctx, Event{Type: "ProxyRequest", Payload: req.URL.String()})
					return
				}
			}
		}
	}

	proxy.handler = &httputil.ReverseProxy{Director: director}

	// Compile route regexes
	for _, service := range conf.Services {
		for _, route := range service.Routes {
			if route.Regex != "" {
				re, err := regexp.Compile(route.Regex)
				if err != nil {
					return nil, err
				}
				proxy.routes = append(proxy.routes, re)
			}
		}
		// Load middleware
		for _, mn := range service.Middleware {
			if mw, ok := registry.Get(mn); ok {
				proxy.middleware = append(proxy.middleware, mw)
			}
		}
	}

	return proxy, nil
}

// Name returns the name of the proxy
func (p *Proxy) Name() string {
	return p.config.Name
}

// Init initializes the proxy with the given config
func (p *Proxy) Init(ctx context.Context, conf *ProxyConfig) error {
	newProxy, err := NewProxy(ctx, conf, p.eventBus, p.registry)
	if err != nil {
		return err
	}
	*p = *newProxy
	return nil
}

// GetHandler returns the http.Handler for the proxy
func (p *Proxy) GetHandler() Handler {
	return p.ApplyMiddleware(p.handler)
}

// GetConfig returns the configuration of the proxy
func (p *Proxy) GetConfig() *ProxyConfig {
	return p.config
}

// MatchRoute checks if the request matches any of the proxy's routes
func (p *Proxy) MatchRoute(r *http.Request) bool {
	for _, service := range p.config.Services {
		for _, route := range service.Routes {
			if matchRoute(r, route) {
				return true
			}
		}
	}
	return false
}

// ApplyMiddleware applies all middleware to the handler
func (p *Proxy) ApplyMiddleware(next http.Handler) http.Handler {
	for i := len(p.middleware) - 1; i >= 0; i-- {
		next = p.middleware[i](next)
	}
	return next
}

func matchRoute(r *http.Request, route RouteMatch) bool {
	// Check path
	if route.Path != "" && !strings.HasPrefix(r.URL.Path, route.Path) {
		return false
	}

	// Check regex
	if route.Regex != "" {
		re, err := regexp.Compile(route.Regex)
		if err != nil || !re.MatchString(r.URL.Path) {
			return false
		}
	}

	// Check methods
	if len(route.Methods) > 0 {
		methodMatch := false
		for _, method := range route.Methods {
			if r.Method == method {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	return true
}
