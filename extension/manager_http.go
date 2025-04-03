package extension

import (
	"fmt"
	"ncobase/ncore/ecode"
	"ncobase/ncore/resp"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ManageRoutes manages routes for all extensions / plugins
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	r.GET("/exts", func(c *gin.Context) {
		extensions := m.GetExtensions()
		result := make(map[string]map[string][]Metadata)

		for _, extension := range extensions {
			group := extension.Metadata.Group
			if group == "" {
				group = extension.Metadata.Name
			}
			if _, ok := result[group]; !ok {
				result[group] = make(map[string][]Metadata)
			}
			result[group][extension.Metadata.Type] = append(result[group][extension.Metadata.Type], extension.Metadata)
		}

		resp.Success(c.Writer, result)
	})

	r.POST("/exts/load", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		fc := m.conf.Extension
		fp := filepath.Join(fc.Path, name+GetPlatformExt())
		if err := m.loadPlugin(fp); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to load extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s loaded successfully", name))
	})

	r.POST("/exts/unload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.UnloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to unload extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s unloaded successfully", name))
	})

	r.POST("/exts/reload", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("name")))
			return
		}
		if err := m.ReloadPlugin(name); err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to reload extension %s: %v", name, err)))
			return
		}
		resp.Success(c.Writer, fmt.Sprintf("%s reloaded successfully", name))
	})
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
func (m *Manager) registerExtensionRoutes(router *gin.Engine, f *Wrapper) {
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
	// log.Infof(context.Background(), "Registered routes for %s with circuit breaker", f.Metadata.Name)
}
