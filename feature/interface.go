package feature

import (
	"ncobase/common/config"

	"github.com/gin-gonic/gin"
)

// Handler represents the handler for a feature
type Handler any

// Service represents the service for a feature
type Service any

// Interface defines the structure for a feature (Plugin / Module)
type Interface interface {
	// Name returns the name of the feature
	Name() string
	// PreInit performs any necessary setup before initialization
	PreInit() error
	// Init initializes the feature with the given config
	Init(conf *config.Config, m *Manager) error
	// PostInit performs any necessary setup after initialization
	PostInit() error
	// RegisterRoutes registers routes for the feature (optional)
	RegisterRoutes(router *gin.RouterGroup)
	// GetHandlers returns the handlers for the feature
	GetHandlers() Handler
	// GetServices returns the services for the feature
	GetServices() Service
	// PreCleanup performs any necessary cleanup before the main cleanup
	PreCleanup() error
	// Cleanup cleans up the feature
	Cleanup() error
	// Status returns the status of the feature
	Status() string
	// GetMetadata returns the metadata of the feature
	GetMetadata() Metadata
	// Version returns the version of the feature
	Version() string
	// Dependencies returns the dependencies of the feature
	Dependencies() []string
}

// Metadata represents the metadata of a feature
type Metadata struct {
	// Name is the name of the feature
	Name string `json:"name,omitempty"`
	// Version is the version of the feature
	Version string `json:"version,omitempty"`
	// Dependencies are the dependencies of the feature
	Dependencies []string `json:"dependencies,omitempty"`
	// Description is the description of the feature
	Description string `json:"description,omitempty"`
	// Group is the belong group of the feature
	Group string `json:"group,omitempty"`
}

// Wrapper wraps a Interface instance with its metadata
type Wrapper struct {
	// Metadata is the metadata of the feature
	Metadata Metadata `json:"metadata"`
	// Instance is the instance of the feature
	Instance Interface `json:"instance,omitempty"`
}
