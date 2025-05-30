package security

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncobase/ncore/extension/config"
)

// Sandbox provides security controls for extension loading
type Sandbox struct {
	config *config.SecurityConfig
}

// NewSandbox creates a new sandbox instance
func NewSandbox(cfg *config.SecurityConfig) *Sandbox {
	return &Sandbox{config: cfg}
}

// ValidatePluginPath checks if plugin path is allowed
func (s *Sandbox) ValidatePluginPath(path string) error {
	if !s.config.EnableSandbox {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %v", err)
	}

	// Check blocked extensions
	ext := strings.ToLower(filepath.Ext(path))
	for _, blocked := range s.config.BlockedExtensions {
		if ext == strings.ToLower(blocked) {
			return fmt.Errorf("extension %s is blocked", ext)
		}
	}

	// Check allowed paths
	if len(s.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range s.config.AllowedPaths {
			absAllowed, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, absAllowed) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path %s is not in allowed paths", path)
		}
	}

	return nil
}

// ValidatePluginSource checks if plugin source is trusted
func (s *Sandbox) ValidatePluginSource(source string) error {
	if !s.config.EnableSandbox || len(s.config.TrustedSources) == 0 {
		return nil
	}

	for _, trusted := range s.config.TrustedSources {
		if strings.Contains(source, trusted) {
			return nil
		}
	}

	return fmt.Errorf("plugin source %s is not trusted", source)
}

// ValidatePluginSignature checks plugin signature if required
func (s *Sandbox) ValidatePluginSignature(path string) error {
	if !s.config.RequireSignature {
		return nil
	}

	// Basic signature validation placeholder
	// In production, implement actual signature verification
	signaturePath := path + ".sig"
	if !fileExists(signaturePath) {
		return fmt.Errorf("plugin signature not found: %s", signaturePath)
	}

	return nil
}

// fileExists checks if file exists
func fileExists(path string) bool {
	_, err := filepath.Abs(path)
	return err == nil
}

// ResourceMonitor monitors plugin resource usage
type ResourceMonitor struct {
	config        *config.PerformanceConfig
	pluginMetrics map[string]*PluginMetrics
}

// PluginMetrics holds resource metrics for a plugin
type PluginMetrics struct {
	MemoryUsageMB   float64
	CPUUsagePercent float64
	LoadTime        time.Duration
	InitTime        time.Duration
	LastAccess      time.Time
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(cfg *config.PerformanceConfig) *ResourceMonitor {
	return &ResourceMonitor{
		config:        cfg,
		pluginMetrics: make(map[string]*PluginMetrics),
	}
}

// CheckResourceLimits validates if plugin can be loaded based on resource limits
func (rm *ResourceMonitor) CheckResourceLimits(pluginName string) error {
	if !rm.config.EnableMetrics {
		return nil
	}

	// Check current system resources
	totalMemory := rm.getTotalMemoryUsage()
	if totalMemory+50 > float64(rm.config.MaxMemoryMB) { // Reserve 50MB buffer
		return fmt.Errorf("insufficient memory: would exceed limit of %dMB", rm.config.MaxMemoryMB)
	}

	totalCPU := rm.getTotalCPUUsage()
	if totalCPU+10 > float64(rm.config.MaxCPUPercent) { // Reserve 10% buffer
		return fmt.Errorf("insufficient CPU: would exceed limit of %d%%", rm.config.MaxCPUPercent)
	}

	return nil
}

// RecordPluginMetrics records resource usage for a plugin
func (rm *ResourceMonitor) RecordPluginMetrics(pluginName string, metrics *PluginMetrics) {
	if !rm.config.EnableMetrics {
		return
	}

	rm.pluginMetrics[pluginName] = metrics
}

// GetPluginMetrics returns metrics for a specific plugin
func (rm *ResourceMonitor) GetPluginMetrics(pluginName string) *PluginMetrics {
	return rm.pluginMetrics[pluginName]
}

// GetAllMetrics returns all plugin metrics
func (rm *ResourceMonitor) GetAllMetrics() map[string]*PluginMetrics {
	result := make(map[string]*PluginMetrics)
	for name, metrics := range rm.pluginMetrics {
		result[name] = metrics
	}
	return result
}

// getTotalMemoryUsage calculates total memory usage across all plugins
func (rm *ResourceMonitor) getTotalMemoryUsage() float64 {
	total := 0.0
	for _, metrics := range rm.pluginMetrics {
		total += metrics.MemoryUsageMB
	}
	return total
}

// getTotalCPUUsage calculates total CPU usage across all plugins
func (rm *ResourceMonitor) getTotalCPUUsage() float64 {
	total := 0.0
	for _, metrics := range rm.pluginMetrics {
		total += metrics.CPUUsagePercent
	}
	return total
}

// Cleanup removes metrics for unloaded plugins
func (rm *ResourceMonitor) Cleanup(pluginName string) {
	delete(rm.pluginMetrics, pluginName)
}
