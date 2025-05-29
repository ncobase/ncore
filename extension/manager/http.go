package manager

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/utils"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ManageRoutes manages routes for all extensions
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	// List all extensions
	r.GET("/exts", func(c *gin.Context) {
		extensions := m.ListExtensions()
		result := make(map[string]map[string][]any)

		for _, ext := range extensions {
			group := ext.Metadata.Group
			if group == "" {
				group = ext.Metadata.Name
			}
			if _, ok := result[group]; !ok {
				result[group] = make(map[string][]any)
			}
			result[group][ext.Metadata.Type] = append(result[group][ext.Metadata.Type], ext.Metadata)
		}

		resp.Success(c.Writer, result)
	})

	// Get extension status
	r.GET("/exts/status", func(c *gin.Context) {
		status := m.GetStatus()
		resp.Success(c.Writer, status)
	})

	// Load an extension
	r.POST("/exts/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest("name is required"))
			return
		}

		fc := m.conf.Extension
		fp := filepath.Join(fc.Path, name+utils.GetPlatformExt())
		if err := m.LoadPlugin(fp); err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to load extension %s: %v", name, err))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s loaded successfully", name))
	})

	// Unload an extension
	r.POST("/exts/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest("name is required"))
			return
		}

		if err := m.UnloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to unload extension %s: %v", name, err))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s unloaded successfully", name))
	})

	// Reload an extension
	r.POST("/exts/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest("name is required"))
			return
		}

		if err := m.ReloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to reload extension %s: %v", name, err))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s reloaded successfully", name))
	})

	// Refresh cross services
	r.POST("/exts/refresh-cross-services", func(c *gin.Context) {
		m.refreshCrossServices()
		resp.Success(c.Writer, "Cross services refreshed successfully")
	})

	// Get comprehensive metrics
	r.GET("/exts/metrics", func(c *gin.Context) {
		metrics := m.GetSystemMetrics()
		resp.Success(c.Writer, metrics)
	})

	// Get specific metric type
	r.GET("/exts/metrics/:type", func(c *gin.Context) {
		metricType := c.Param("type")

		var result any
		switch metricType {
		case "events":
			result = m.GetEventsMetrics()
		case "cache":
			result = m.GetServiceCacheStats()
		case "extensions":
			result = m.GetExtensionMetrics()
		case "system":
			result = m.GetSystemMetrics()
		default:
			resp.Fail(c.Writer, resp.BadRequest("Invalid metric type. Use: events, cache, extensions, system"))
			return
		}

		resp.Success(c.Writer, result)
	})
}

// RegisterRoutes registers all extension routes with the provided router
func (m *Manager) RegisterRoutes(router *gin.Engine) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ext := range m.extensions {
		if ext.Instance.GetHandlers() != nil {
			m.registerExtensionRoutes(router, ext)
		}
	}
}

// registerExtensionRoutes registers routes for a single extension
func (m *Manager) registerExtensionRoutes(router *gin.Engine, ext *types.Wrapper) {
	// Create circuit breaker for this extension
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        ext.Metadata.Name,
		MaxRequests: 100,
		Interval:    5 * time.Second,
		Timeout:     3 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	m.circuitBreakers[ext.Metadata.Name] = cb

	// Register extension routes
	group := router.Group("")
	ext.Instance.RegisterRoutes(group)
}
