package manager

import (
	"fmt"
	"ncore/ext/core"
	"ncore/pkg/utils"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ManageRoutes manages routes for all extensions / plugins
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	r.GET("/exts", func(c *gin.Context) {
		extensions := m.GetExtensions()
		result := make(map[string]map[string][]interface{})

		for _, ext := range extensions {
			group := ext.Metadata.Group
			if group == "" {
				group = ext.Metadata.Name
			}
			if _, ok := result[group]; !ok {
				result[group] = make(map[string][]interface{})
			}
			result[group][ext.Metadata.Type] = append(result[group][ext.Metadata.Type], ext.Metadata)
		}

		// Use your actual response helper here
		c.JSON(200, gin.H{"success": true, "data": result})
	})

	r.POST("/exts/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"success": false, "message": "name is required"})
			return
		}
		fc := m.conf.Extension
		fp := filepath.Join(fc.Path, name+utils.GetPlatformExt())
		if err := m.LoadPlugin(fp); err != nil {
			c.JSON(500, gin.H{"success": false, "message": fmt.Sprintf("Failed to load extension %s: %v", name, err)})
			return
		}
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("%s loaded successfully", name)})
	})

	r.POST("/exts/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"success": false, "message": "name is required"})
			return
		}
		if err := m.UnloadPlugin(name); err != nil {
			c.JSON(500, gin.H{"success": false, "message": fmt.Sprintf("Failed to unload extension %s: %v", name, err)})
			return
		}
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("%s unloaded successfully", name)})
	})

	r.POST("/exts/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"success": false, "message": "name is required"})
			return
		}
		if err := m.ReloadPlugin(name); err != nil {
			c.JSON(500, gin.H{"success": false, "message": fmt.Sprintf("Failed to reload extension %s: %v", name, err)})
			return
		}
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("%s reloaded successfully", name)})
	})
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (m *Manager) ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error) {
	cb, ok := m.circuitBreakers[extensionName]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found for extension %s", extensionName)
	}

	return cb.Execute(fn)
}

// RegisterRoutes registers all extension routes with the provided router
func (m *Manager) RegisterRoutes(router *gin.Engine) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.extensions {
		if f.Instance.GetHandlers() != nil {
			m.registerExtensionRoutes(router, f)
		}
	}
}

// registerExtensionRoutes registers routes for a single extension
func (m *Manager) registerExtensionRoutes(router *gin.Engine, f *core.Wrapper) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        f.Metadata.Name,
		MaxRequests: 100,
		Interval:    5 * time.Second,
		Timeout:     3 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})
	m.circuitBreakers[f.Metadata.Name] = cb
	group := router.Group("")
	f.Instance.RegisterRoutes(group)
}
