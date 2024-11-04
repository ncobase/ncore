package extension

import (
	"ncobase/common/config"

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
	// PreInit performs any necessary setup before initialization
	PreInit() error
	// Init initializes the extension with the given config
	Init(conf *config.Config, m *Manager) error
	// PostInit performs any necessary setup after initialization
	PostInit() error
	// RegisterRoutes registers routes for the extension (optional)
	RegisterRoutes(router *gin.RouterGroup)
	// GetHandlers returns the handlers for the extension
	GetHandlers() Handler
	// GetServices returns the services for the extension
	GetServices() Service
	// PreCleanup performs any necessary cleanup before the main cleanup
	PreCleanup() error
	// Cleanup cleans up the extension
	Cleanup() error
	// Status returns the status of the extension
	Status() string
	// GetMetadata returns the metadata of the extension
	GetMetadata() Metadata
	// Version returns the version of the extension
	Version() string
	// Dependencies returns the dependencies of the extension
	Dependencies() []string
	// NeedServiceDiscovery returns if the extension needs to be registered as a service
	NeedServiceDiscovery() bool
	// GetServiceInfo returns service registration info if NeedServiceDiscovery returns true
	GetServiceInfo() *ServiceInfo
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
	// Type is the type of the extension, e.g. plugin or module
	Type string `json:"type,omitempty"`
	// Group is the belong group of the extension
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
