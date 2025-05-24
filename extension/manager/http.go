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

// ManageRoutes manages routes for all extensions / plugins
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	r.GET("/exts", func(c *gin.Context) {
		extensions := m.GetExtensions()
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

	r.POST("/exts/refresh-cross-services", func(c *gin.Context) {
		m.RefreshCrossServices()
		resp.Success(c.Writer, "Cross services refreshed successfully")
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
func (m *Manager) registerExtensionRoutes(router *gin.Engine, f *types.Wrapper) {
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
