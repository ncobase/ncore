package extension

import (
	"ncobase/ncore/config"

	"github.com/gin-gonic/gin"
)

// Handler represents the handler for a extension
type Handler any

// Service represents the service for a extension
type Service any

// Interface defines the structure for a extension (Plugin / Module)
type Interface interface {
	// Name returns the name of the extension
	Name() string
	// Init initializes the extension with the given config
	Init(conf *config.Config, m *Manager) error
	// GetHandlers returns the handlers for the extension
	GetHandlers() Handler
	// GetServices returns the services for the extension
	GetServices() Service
	// GetMetadata returns the metadata of the extension
	GetMetadata() Metadata
	// Version returns the version of the extension
	Version() string
	// Dependencies returns the dependencies of the extension
	Dependencies() []string
	// OptionalMethods returns the optional methods of the extension
	OptionalMethods
}

// OptionalMethods represents the optional methods for a extension
type OptionalMethods interface {
	// PreInit performs any necessary setup before initialization
	PreInit() error
	// PostInit performs any necessary setup after initialization
	PostInit() error
	// RegisterRoutes registers routes for the extension (optional)
	RegisterRoutes(router *gin.RouterGroup)
	// PreCleanup performs any necessary cleanup before the main cleanup
	PreCleanup() error
	// Cleanup cleans up the extension
	Cleanup() error
	// Status returns the status of the extension
	Status() string
	// NeedServiceDiscovery returns if the extension needs to be registered as a service
	NeedServiceDiscovery() bool
	// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
	GetServiceInfo() *ServiceInfo
}

// Extension status constants
const (
	// StatusActive indicates the extension is running normally
	StatusActive = "active"
	// StatusInactive indicates the extension is installed but not running
	StatusInactive = "inactive"
	// StatusError indicates the extension encountered an error
	StatusError = "error"
	// StatusInitializing indicates the extension is in initialization process
	StatusInitializing = "initializing"
	// StatusMaintenance indicates the extension is under maintenance
	StatusMaintenance = "maintenance"
	// StatusDisabled indicates the extension has been manually disabled
	StatusDisabled = "disabled"
)

// OptionalImpl implements the optional methods
type OptionalImpl struct{}

// PreInit performs any necessary setup before initialization
func (o *OptionalImpl) PreInit() error {
	return nil
}

// PostInit performs any necessary setup after initialization
func (o *OptionalImpl) PostInit() error {
	return nil
}

// RegisterRoutes registers routes for the extension
func (o *OptionalImpl) RegisterRoutes(router *gin.RouterGroup) {}

// PreCleanup performs any necessary cleanup before the main cleanup
func (o *OptionalImpl) PreCleanup() error {
	return nil
}

// Cleanup cleans up the extension
func (o *OptionalImpl) Cleanup() error {
	return nil
}

// Status returns the status of the extension
func (o *OptionalImpl) Status() string {
	return StatusActive
}

// NeedServiceDiscovery returns if the extension needs to be registered as a service
func (o *OptionalImpl) NeedServiceDiscovery() bool {
	return false
}

// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
func (o *OptionalImpl) GetServiceInfo() *ServiceInfo {
	return nil
}

// Metadata represents the metadata of a extension
type Metadata struct {
	// Name is the name of the extension
	Name string `json:"name,omitempty"`
	// Version is the version of the extension
	Version string `json:"version,omitempty"`
	// Dependencies are the dependencies of the extension
	Dependencies []string `json:"dependencies,omitempty"`
	// Description is the description of the extension
	Description string `json:"description,omitempty"`
	// Type is the type of the extension, e.g. core, bisiness, plugin, module, etc
	Type string `json:"type,omitempty"`
	// Group is the belong group of the extension, e.g. iam, res, flow, sys, org, rt, plug, etc
	Group string `json:"group,omitempty"`
}

// Wrapper wraps a Interface instance with its metadata
type Wrapper struct {
	// Metadata is the metadata of the extension
	Metadata Metadata `json:"metadata"`
	// Instance is the instance of the extension
	Instance Interface `json:"instance,omitempty"`
}

// ServiceInfo contains information needed for service registration
type ServiceInfo struct {
	Address string            `json:"address"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
}
