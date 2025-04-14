package types

// Service Status constants
const (
	ServiceStatusHealthy   = "healthy"
	ServiceStatusUnhealthy = "unhealthy"
	ServiceStatusUnknown   = "unknown"
)

// ServiceInfo contains information needed for service registration
type ServiceInfo struct {
	Address string            `json:"address"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
}
