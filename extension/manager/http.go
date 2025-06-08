package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ncobase/ncore/extension/metrics"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/utils"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ManageRoutes manages routes for all extensions
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	// Extension management routes
	m.setupExtensionRoutes(r)

	// Metrics routes
	m.setupMetricsRoutes(r)

	// Health check routes
	m.setupHealthRoutes(r)
}

// setupExtensionRoutes sets up extension management routes
func (m *Manager) setupExtensionRoutes(r *gin.RouterGroup) {
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

	// Plugin management routes
	pluginGroup := r.Group("/exts/plugins")
	{
		pluginGroup.POST("/load", func(c *gin.Context) {
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

		pluginGroup.POST("/unload", func(c *gin.Context) {
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

		pluginGroup.POST("/reload", func(c *gin.Context) {
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
	}

	// Cross services management
	r.POST("/exts/cross-services/refresh", func(c *gin.Context) {
		m.refreshCrossServices()
		resp.Success(c.Writer, "Cross services refreshed successfully")
	})
}

// setupMetricsRoutes sets up metrics routes
func (m *Manager) setupMetricsRoutes(r *gin.RouterGroup) {
	metricsGroup := r.Group("/exts/metrics")
	{
		// Main metrics endpoint
		metricsGroup.GET("", func(c *gin.Context) {
			result := m.GetMetrics()
			resp.Success(c.Writer, result)
		})

		// Historical queries
		metricsGroup.GET("/history", func(c *gin.Context) {
			opts, err := parseQueryOptions(c)
			if err != nil {
				resp.Fail(c.Writer, resp.BadRequest("Invalid query parameters: %v", err))
				return
			}

			results, err := m.QueryHistoricalMetrics(opts)
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Query failed: %v", err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"query":   opts,
				"results": results,
				"count":   len(results),
			})
		})

		// Latest metrics by extension
		metricsGroup.GET("/latest/:name", func(c *gin.Context) {
			name := c.Param("name")
			limitStr := c.DefaultQuery("limit", "100")

			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit <= 0 || limit > 1000 {
				limit = 100
			}

			snapshots, err := m.GetLatestMetrics(name, limit)
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Failed to get latest metrics: %v", err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"extension": name,
				"limit":     limit,
				"count":     len(snapshots),
				"snapshots": snapshots,
			})
		})

		// Specific extension metrics
		metricsGroup.GET("/extensions/:name", func(c *gin.Context) {
			name := c.Param("name")
			result := m.GetExtensionMetrics(name)

			if result == nil {
				resp.Fail(c.Writer, resp.NotFound("Extension '%s' not found", name))
				return
			}

			resp.Success(c.Writer, result)
		})

		// Events metrics
		metricsGroup.GET("/events", func(c *gin.Context) {
			eventMetrics := m.GetEventsMetrics()
			resp.Success(c.Writer, eventMetrics)
		})

		// Storage info
		metricsGroup.GET("/storage", func(c *gin.Context) {
			stats := m.GetMetricsStorageStats()
			resp.Success(c.Writer, stats)
		})
	}
}

// setupHealthRoutes sets up simplified health check routes
func (m *Manager) setupHealthRoutes(r *gin.RouterGroup) {
	healthGroup := r.Group("/exts/health")
	{
		// Overall system health
		healthGroup.GET("", func(c *gin.Context) {
			health := m.buildSystemHealth(c.Request.Context())

			status := health["status"].(string)
			if status == "healthy" {
				resp.Success(c.Writer, health)
			} else {
				c.Writer.WriteHeader(503)
				resp.Success(c.Writer, health)
			}
		})

		// Extension health summary
		healthGroup.GET("/extensions", func(c *gin.Context) {
			extensionStatus := m.GetStatus()
			summary := map[string]int{
				"total":  len(extensionStatus),
				"active": 0,
				"error":  0,
				"other":  0,
			}

			for _, status := range extensionStatus {
				switch status {
				case types.StatusActive:
					summary["active"]++
				case types.StatusError:
					summary["error"]++
				default:
					summary["other"]++
				}
			}

			resp.Success(c.Writer, map[string]any{
				"summary":    summary,
				"extensions": extensionStatus,
			})
		})

		// Data layer health
		healthGroup.GET("/data", func(c *gin.Context) {
			if m.data == nil {
				resp.Fail(c.Writer, resp.ServiceUnavailable("Data layer not initialized"))
				return
			}

			health := m.data.Health(c.Request.Context())
			status := health["status"].(string)

			if status == "healthy" {
				resp.Success(c.Writer, health)
			} else {
				resp.Fail(c.Writer, resp.ServiceUnavailable("Data layer unhealthy: %v", health))
			}
		})

		// Circuit breaker status
		healthGroup.GET("/circuit-breakers", func(c *gin.Context) {
			breakerStatus := m.getCircuitBreakerStatus()
			resp.Success(c.Writer, breakerStatus)
		})
	}
}

// parseQueryOptions parses query parameters into QueryOptions
func parseQueryOptions(c *gin.Context) (*metrics.QueryOptions, error) {
	opts := &metrics.QueryOptions{
		ExtensionName: c.Query("extension"),
		MetricType:    c.Query("metric_type"),
		Aggregation:   c.DefaultQuery("aggregation", "raw"),
		Limit:         100,
	}

	// Parse time range
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		// Default to last hour
		opts.EndTime = time.Now()
		opts.StartTime = opts.EndTime.Add(-time.Hour)
	} else {
		var err error
		opts.StartTime, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start time format: %v", err)
		}

		opts.EndTime, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end time format: %v", err)
		}
	}

	// Parse interval
	if intervalStr := c.Query("interval"); intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid interval format: %v", err)
		}
		opts.Interval = interval
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 10000 {
			return nil, fmt.Errorf("invalid limit: must be 1-10000")
		}
		opts.Limit = limit
	}

	return opts, nil
}

// buildSystemHealth builds comprehensive system health status
func (m *Manager) buildSystemHealth(ctx context.Context) map[string]any {
	health := map[string]any{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"extensions": len(m.extensions),
		"components": make(map[string]any),
	}

	components := health["components"].(map[string]any)
	overallHealthy := true

	// Extension health
	extensionStatus := m.GetStatus()
	healthyExts := 0
	for _, status := range extensionStatus {
		if status == types.StatusActive {
			healthyExts++
		}
	}

	extHealthRate := float64(healthyExts) / float64(len(extensionStatus)) * 100
	components["extensions"] = map[string]any{
		"total":   len(extensionStatus),
		"healthy": healthyExts,
		"rate":    extHealthRate,
		"status":  getHealthStatus(extHealthRate),
	}

	if healthyExts < len(extensionStatus) {
		overallHealthy = false
	}

	// Data layer health
	if m.data != nil {
		dataHealth := m.data.Health(ctx)
		components["data"] = dataHealth

		if status, ok := dataHealth["status"].(string); ok && status != "healthy" {
			overallHealthy = false
		}
	}

	// Metrics system health
	if m.metricsCollector != nil && m.metricsCollector.IsEnabled() {
		components["metrics"] = map[string]any{
			"status": "enabled",
			"stats":  m.GetMetricsStorageStats(),
		}
	} else {
		components["metrics"] = map[string]any{
			"status": "disabled",
		}
	}

	// Service discovery health
	if m.serviceDiscovery != nil {
		cacheStats := m.GetServiceCacheStats()
		components["service_discovery"] = map[string]any{
			"status": "enabled",
			"cache":  cacheStats,
		}
	} else {
		components["service_discovery"] = map[string]any{
			"status": "disabled",
		}
	}

	if !overallHealthy {
		health["status"] = "degraded"
	}

	return health
}

// getHealthStatus returns health status based on percentage
func getHealthStatus(percentage float64) string {
	switch {
	case percentage >= 95:
		return "excellent"
	case percentage >= 80:
		return "good"
	case percentage >= 60:
		return "degraded"
	default:
		return "unhealthy"
	}
}

// getCircuitBreakerStatus returns circuit breaker status
func (m *Manager) getCircuitBreakerStatus() map[string]any {
	if len(m.circuitBreakers) == 0 {
		return map[string]any{
			"total":  0,
			"status": "no_circuit_breakers",
		}
	}

	breakerStatus := make(map[string]any)
	totalBreakers := len(m.circuitBreakers)
	openBreakers := 0

	for name, cb := range m.circuitBreakers {
		state := cb.State().String()
		counts := cb.Counts()

		breakerStatus[name] = map[string]any{
			"state":           state,
			"requests":        counts.Requests,
			"total_successes": counts.TotalSuccesses,
			"total_failures":  counts.TotalFailures,
		}

		if cb.State() == gobreaker.StateOpen {
			openBreakers++
		}
	}

	return map[string]any{
		"total":    totalBreakers,
		"open":     openBreakers,
		"closed":   totalBreakers - openBreakers,
		"breakers": breakerStatus,
	}
}

// RegisterRoutes registers all extension routes
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
