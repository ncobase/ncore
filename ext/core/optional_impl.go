package core

import "github.com/gin-gonic/gin"

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
