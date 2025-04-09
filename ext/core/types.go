package core

import "time"

// ServiceInfo contains information needed for service registration
type ServiceInfo struct {
	Address string            `json:"address"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
}

// Metadata represents the metadata of an extension
type Metadata struct {
	// Name is the name of the extension
	Name string `json:"name,omitempty"`
	// Version is the version of the extension
	Version string `json:"version,omitempty"`
	// Dependencies are the dependencies of the extension
	Dependencies []string `json:"dependencies,omitempty"`
	// Description is the description of the extension
	Description string `json:"description,omitempty"`
	// Type is the type of the extension, e.g. core, business, plugin, module, etc
	Type string `json:"type,omitempty"`
	// Group is the belong group of the extension, e.g. iam, res, flow, sys, org, rt, plug, etc
	Group string `json:"group,omitempty"`
}

// EventData basic event data
type EventData struct {
	Time      time.Time
	Source    string
	EventType string
	Data      any
}

// Service Status constants
const (
	ServiceStatusHealthy   = "healthy"
	ServiceStatusUnhealthy = "unhealthy"
	ServiceStatusUnknown   = "unknown"
)

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
