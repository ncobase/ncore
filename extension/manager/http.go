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
	// Create API group
	apiGroup := r.Group("")

	// Extension management routes
	m.setupExtensionRoutes(apiGroup)

	// Plugin management routes
	m.setupPluginRoutes(apiGroup)

	// Metrics routes
	m.setupMetricsRoutes(apiGroup)

	// Health check routes
	m.setupHealthRoutes(apiGroup)

	// System management routes
	m.setupSystemRoutes(apiGroup)
}

// setupExtensionRoutes sets up extension management routes
func (m *Manager) setupExtensionRoutes(r *gin.RouterGroup) {
	extGroup := r.Group("/extensions")
	{
		// List all extensions
		extGroup.GET("", func(c *gin.Context) {
			extensions := m.ListExtensions()

			// Group by category for better organization
			result := make(map[string]map[string][]any)
			for _, ext := range extensions {
				group := ext.Metadata.Group
				if group == "" {
					group = "default"
				}

				extType := ext.Metadata.Type
				if extType == "" {
					extType = "module"
				}

				if _, ok := result[group]; !ok {
					result[group] = make(map[string][]any)
				}

				result[group][extType] = append(result[group][extType], ext.Metadata)
			}

			resp.Success(c.Writer, result)
		})

		// Get extension status
		extGroup.GET("/status", func(c *gin.Context) {
			status := m.GetStatus()

			// Add summary information
			summary := map[string]int{
				"total":  len(status),
				"active": 0,
				"error":  0,
				"other":  0,
			}

			for _, s := range status {
				switch s {
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
				"extensions": status,
			})
		})

		// Get specific extension info
		extGroup.GET("/:name", func(c *gin.Context) {
			name := c.Param("name")
			ext, err := m.GetExtensionByName(name)
			if err != nil {
				resp.Fail(c.Writer, resp.NotFound("Extension '%s' not found", name))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"metadata": ext.GetMetadata(),
				"status":   ext.Status(),
			})
		})

		// Get extension metadata
		extGroup.GET("/metadata", func(c *gin.Context) {
			metadata := m.GetMetadata()
			resp.Success(c.Writer, metadata)
		})
	}
}

// setupPluginRoutes sets up plugin management routes
func (m *Manager) setupPluginRoutes(r *gin.RouterGroup) {
	pluginGroup := r.Group("/plugins")
	{
		// Load plugin
		pluginGroup.POST("/load", func(c *gin.Context) {
			name := c.Query("name")
			if name == "" {
				resp.Fail(c.Writer, resp.BadRequest("Plugin name is required"))
				return
			}

			fc := m.conf.Extension
			fp := filepath.Join(fc.Path, name+utils.GetPlatformExt())

			if err := m.LoadPlugin(fp); err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Failed to load plugin %s: %v", name, err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"message": fmt.Sprintf("Plugin %s loaded successfully", name),
				"plugin":  name,
			})
		})

		// Unload plugin
		pluginGroup.POST("/unload", func(c *gin.Context) {
			name := c.Query("name")
			if name == "" {
				resp.Fail(c.Writer, resp.BadRequest("Plugin name is required"))
				return
			}

			if err := m.UnloadPlugin(name); err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Failed to unload plugin %s: %v", name, err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"message": fmt.Sprintf("Plugin %s unloaded successfully", name),
				"plugin":  name,
			})
		})

		// Reload plugin
		pluginGroup.POST("/reload", func(c *gin.Context) {
			name := c.Query("name")
			if name == "" {
				resp.Fail(c.Writer, resp.BadRequest("Plugin name is required"))
				return
			}

			if err := m.ReloadPlugin(name); err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Failed to reload plugin %s: %v", name, err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"message": fmt.Sprintf("Plugin %s reloaded successfully", name),
				"plugin":  name,
			})
		})
	}
}

// setupMetricsRoutes sets up metrics routes
func (m *Manager) setupMetricsRoutes(r *gin.RouterGroup) {
	metricsGroup := r.Group("/metrics")
	{
		// Dashboard summary
		metricsGroup.GET("/summary", func(c *gin.Context) {
			summary := map[string]any{
				"timestamp":         time.Now(),
				"total_extensions":  len(m.extensions),
				"active_extensions": m.countActiveExtensions(),
				"data_status":       m.getDataLayerStatus(),
				"messaging":         m.getMessagingStatus(),
				"metrics_enabled":   m.metricsCollector != nil && m.metricsCollector.IsEnabled(),
			}

			if m.serviceDiscovery != nil {
				cacheStats := m.GetServiceCacheStats()
				summary["service_discovery"] = map[string]any{
					"enabled":    true,
					"cache_hits": cacheStats["cache_hits"],
					"hit_rate":   cacheStats["hit_rate"],
				}
			}

			resp.Success(c.Writer, summary)
		})

		// System-wide metrics
		metricsGroup.GET("/system", func(c *gin.Context) {
			result := m.GetSystemMetrics()
			resp.Success(c.Writer, result)
		})

		// Comprehensive metrics
		metricsGroup.GET("/comprehensive", func(c *gin.Context) {
			result := m.GetComprehensiveMetrics()
			resp.Success(c.Writer, result)
		})

		// Extension metrics
		metricsGroup.GET("/extensions", func(c *gin.Context) {
			result := m.GetMetrics()
			resp.Success(c.Writer, result)
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

		// Data layer metrics
		metricsGroup.GET("/data", func(c *gin.Context) {
			result := m.GetDataMetrics()
			resp.Success(c.Writer, result)
		})

		// Events metrics
		metricsGroup.GET("/events", func(c *gin.Context) {
			eventMetrics := m.GetEventsMetrics()
			resp.Success(c.Writer, eventMetrics)
		})

		// Service discovery metrics
		metricsGroup.GET("/service-discovery", func(c *gin.Context) {
			cacheStats := m.GetServiceCacheStats()
			resp.Success(c.Writer, cacheStats)
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

		// Latest metrics
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

		// Storage info
		metricsGroup.GET("/storage", func(c *gin.Context) {
			stats := m.GetMetricsStorageStats()
			resp.Success(c.Writer, stats)
		})
	}
}

// setupHealthRoutes sets up health check routes
func (m *Manager) setupHealthRoutes(r *gin.RouterGroup) {
	healthGroup := r.Group("/health")
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

		// Extension health
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
				c.Writer.WriteHeader(503)
				resp.Success(c.Writer, health)
			}
		})

		// Circuit breaker status
		healthGroup.GET("/circuit-breakers", func(c *gin.Context) {
			breakerStatus := m.getCircuitBreakerStatus()
			resp.Success(c.Writer, breakerStatus)
		})
	}
}

// setupSystemRoutes sets up system management routes
func (m *Manager) setupSystemRoutes(r *gin.RouterGroup) {
	systemGroup := r.Group("/system")
	{
		// System info
		systemGroup.GET("/info", func(c *gin.Context) {
			info := map[string]any{
				"version":    "1.0.0", // You can inject this from build
				"build_time": time.Now().Format(time.RFC3339),
				"uptime":     time.Since(time.Now()).String(), // Calculate actual uptime
				"extensions": map[string]any{
					"total":      len(m.extensions),
					"active":     m.countActiveExtensions(),
					"hot_reload": m.conf.Extension.HotReload,
				},
			}

			resp.Success(c.Writer, info)
		})

		// Cross services management
		systemGroup.POST("/cross-services/refresh", func(c *gin.Context) {
			m.refreshCrossServices()
			resp.Success(c.Writer, map[string]any{
				"message": "Cross services refreshed successfully",
			})
		})

		// Configuration info (non-sensitive parts)
		systemGroup.GET("/config", func(c *gin.Context) {
			config := map[string]any{
				"extension": map[string]any{
					"mode":        m.conf.Extension.Mode,
					"path":        m.conf.Extension.Path,
					"hot_reload":  m.conf.Extension.HotReload,
					"max_plugins": m.conf.Extension.MaxPlugins,
				},
				"features": map[string]any{
					"grpc_enabled":    m.conf.GRPC != nil && m.conf.GRPC.Enabled,
					"consul_enabled":  m.conf.Consul != nil,
					"metrics_enabled": m.metricsCollector != nil && m.metricsCollector.IsEnabled(),
				},
			}

			resp.Success(c.Writer, config)
		})
	}
}

// Helper functions remain the same as before...

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

// registerExtensionRoutes registers routes for a single extension with circuit breaker
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
