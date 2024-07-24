package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Manager represents the manager for proxies
type Manager struct {
	proxies  map[string]Interface
	mu       sync.RWMutex
	eventBus EventBusInterface
	registry RegistryInterface
}

// NewManager creates a new proxy manager
func NewManager(eventBus EventBusInterface, registry RegistryInterface) *Manager {
	return &Manager{
		proxies:  make(map[string]Interface),
		eventBus: eventBus,
		registry: registry,
	}
}

// Register registers a new proxy
func (pm *Manager) Register(p Interface) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.proxies[p.Name()]; exists {
		return fmt.Errorf("proxy %s already registered", p.Name())
	}
	pm.proxies[p.Name()] = p
	pm.eventBus.Publish(context.Background(), Event{Type: "ProxyRegistered", Payload: p.Name()})
	return nil
}

// GetProxy returns a registered proxy by name
func (pm *Manager) GetProxy(name string) (Interface, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.proxies[name]
	if !exists {
		return nil, fmt.Errorf("proxy %s not found", name)
	}
	return p, nil
}

// ListProxies returns a list of all registered proxies
func (pm *Manager) ListProxies() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var names []string
	for name := range pm.proxies {
		names = append(names, name)
	}
	return names
}

// RemoveProxy removes a registered proxy
func (pm *Manager) RemoveProxy(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.proxies[name]; !exists {
		return fmt.Errorf("proxy %s not found", name)
	}
	delete(pm.proxies, name)
	pm.eventBus.Publish(context.Background(), Event{Type: "ProxyRemoved", Payload: name})
	return nil
}

// ServeHTTP implements the http.Handler interface
func (pm *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, proxy := range pm.proxies {
		if proxy.MatchRoute(r) {
			proxy.GetHandler().ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// HandleProxyAPI handles the proxy management API
func (pm *Manager) HandleProxyAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pm.listProxies(w, r)
	case http.MethodPost:
		pm.createProxy(w, r)
	case http.MethodPut:
		pm.updateProxy(w, r)
	case http.MethodDelete:
		pm.deleteProxy(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (pm *Manager) listProxies(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(pm.ListProxies())
	if err != nil {
		return
	}
}

func (pm *Manager) createProxy(w http.ResponseWriter, r *http.Request) {
	var config ProxyConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proxy, err := NewProxy(r.Context(), &config, pm.eventBus, pm.registry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := pm.Register(proxy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(config)
	if err != nil {
		return
	}
}

func (pm *Manager) updateProxy(w http.ResponseWriter, r *http.Request) {
	var config ProxyConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proxy, err := pm.GetProxy(config.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := proxy.Init(r.Context(), &config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pm.eventBus.Publish(r.Context(), Event{Type: "ProxyUpdated", Payload: config.Name})
	err = json.NewEncoder(w).Encode(config)
	if err != nil {
		return
	}
}

func (pm *Manager) deleteProxy(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Proxy name is required", http.StatusBadRequest)
		return
	}

	if err := pm.RemoveProxy(name); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
