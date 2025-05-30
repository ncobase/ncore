package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/utils"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ManageRoutes manages routes for all extensions
func (m *Manager) ManageRoutes(r *gin.RouterGroup) {
	// Core extension management routes
	m.setupCoreRoutes(r)

	// Metrics-specific routes
	m.setupMetricsRoutes(r)

	// Health check routes
	m.setupHealthRoutes(r)
}

// setupCoreRoutes sets up core extension management routes
func (m *Manager) setupCoreRoutes(r *gin.RouterGroup) {
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

	// Cross services management
	r.POST("/exts/refresh-cross-services", func(c *gin.Context) {
		m.refreshCrossServices()
		resp.Success(c.Writer, "Cross services refreshed successfully")
	})
}

// setupMetricsRoutes sets up metrics-related routes
func (m *Manager) setupMetricsRoutes(r *gin.RouterGroup) {
	// Get comprehensive metrics
	r.GET("/exts/metrics", func(c *gin.Context) {
		metrics := m.GetMetrics()
		resp.Success(c.Writer, metrics)
	})

	// Get specific metric types
	r.GET("/exts/metrics/:type", func(c *gin.Context) {
		metricType := c.Param("type")
		result := m.GetSpecificMetrics(metricType)

		if errorMsg, hasError := result["error"]; hasError && errorMsg != "" {
			validTypes := []string{"collections", "storage", "service_cache", "cache", "data", "system", "security", "resource"}
			resp.Fail(c.Writer, resp.BadRequest("Invalid metric type '%s'. Valid types: %v", metricType, validTypes))
			return
		}

		resp.Success(c.Writer, result)
	})

	// Metrics collections operations
	r.GET("/exts/metrics/collections", func(c *gin.Context) {
		if m.metricsManager == nil || !m.metricsManager.IsEnabled() {
			resp.Success(c.Writer, map[string]any{
				"collections": []string{},
				"details":     map[string]any{},
			})
			return
		}

		collections := m.metricsManager.GetAllCollections()
		collectionNames := make([]string, 0, len(collections))
		collectionInfo := make(map[string]any)

		for name, collection := range collections {
			collectionNames = append(collectionNames, name)
			collectionInfo[name] = map[string]any{
				"metric_count": len(collection.Metrics),
				"last_updated": collection.LastUpdated,
			}
		}

		resp.Success(c.Writer, map[string]any{
			"collections": collectionNames,
			"details":     collectionInfo,
		})
	})

	// Get specific collection data
	r.GET("/exts/metrics/collections/:collection", func(c *gin.Context) {
		collection := c.Param("collection")

		if m.metricsManager == nil || !m.metricsManager.IsEnabled() {
			resp.Fail(c.Writer, resp.ServiceUnavailable("Metrics collector not initialized"))
			return
		}

		collectionData, exists := m.metricsManager.GetAllCollections()[collection]
		if !exists {
			resp.Fail(c.Writer, resp.NotFound("Collection '%s' not found", collection))
			return
		}

		snapshots := make([]map[string]any, 0, len(collectionData.Metrics))
		for _, metric := range collectionData.Metrics {
			snapshots = append(snapshots, map[string]any{
				"name":      metric.Name,
				"type":      m.getMetricTypeString(metric.Type),
				"value":     metric.Value.Load(),
				"labels":    metric.Labels,
				"timestamp": metric.Timestamp,
				"help":      metric.Help,
				"unit":      metric.Unit,
			})
		}

		resp.Success(c.Writer, map[string]any{
			"collection":   collection,
			"metrics":      snapshots,
			"last_updated": collectionData.LastUpdated,
		})
	})

	// Query historical metrics
	r.GET("/exts/metrics/query/:collection", func(c *gin.Context) {
		collection := c.Param("collection")
		startStr := c.Query("start")
		endStr := c.Query("end")
		limitStr := c.Query("limit")

		if limitStr != "" {
			limit := 100
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
				limit = parsed
			}

			snapshots, err := m.GetLatestMetrics(collection, limit)
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer("Failed to query metrics: %v", err))
				return
			}

			resp.Success(c.Writer, map[string]any{
				"collection": collection,
				"limit":      limit,
				"count":      len(snapshots),
				"data":       snapshots,
			})
			return
		}

		if startStr == "" || endStr == "" {
			resp.Fail(c.Writer, resp.BadRequest("start and end parameters required for time range query, or use limit for latest metrics"))
			return
		}

		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid start time format, use RFC3339"))
			return
		}

		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid end time format, use RFC3339"))
			return
		}

		snapshots, err := m.QueryMetrics(collection, start, end)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to query metrics: %v", err))
			return
		}

		resp.Success(c.Writer, map[string]any{
			"collection": collection,
			"start":      start,
			"end":        end,
			"count":      len(snapshots),
			"data":       snapshots,
		})
	})

	// Get metrics snapshot
	r.GET("/exts/metrics/snapshot", func(c *gin.Context) {
		if m.metricsManager == nil || !m.metricsManager.IsEnabled() {
			resp.Fail(c.Writer, resp.ServiceUnavailable("Metrics collector not initialized"))
			return
		}

		snapshot := m.metricsManager.Snapshot()
		resp.Success(c.Writer, snapshot)
	})

	// Metrics storage operations
	r.GET("/exts/metrics/storage", func(c *gin.Context) {
		stats := m.GetMetricsStats()
		resp.Success(c.Writer, stats)
	})
}

// setupHealthRoutes sets up health check routes
func (m *Manager) setupHealthRoutes(r *gin.RouterGroup) {
	// Data layer health
	r.GET("/exts/health/data", func(c *gin.Context) {
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

	// Overall system health
	r.GET("/exts/health", func(c *gin.Context) {
		health := m.buildSystemHealth(c.Request.Context())

		status := health["status"].(string)
		if status == "healthy" {
			resp.Success(c.Writer, health)
		} else {
			c.Writer.WriteHeader(503)
			resp.Success(c.Writer, health)
		}
	})

	// Circuit breaker status
	r.GET("/exts/health/circuit-breakers", func(c *gin.Context) {
		breakerStatus := m.getCircuitBreakerStatus()
		resp.Success(c.Writer, breakerStatus)
	})

	// Extension status and summary
	r.GET("/exts/health/extensions", func(c *gin.Context) {
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
	if m.metricsManager != nil && m.metricsManager.IsEnabled() {
		components["metrics"] = map[string]any{
			"status": "enabled",
			"stats":  m.GetMetricsStats(),
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
