// Package handler provides HTTP handlers for the basic REST API example.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/service"
	"github.com/ncobase/ncore/logging/logger"
)

// Handler aggregates all HTTP handlers.
type Handler struct {
	Task   *TaskHandler
	logger *logger.Logger
}

// NewHandler creates a new handler instance with all sub-handlers initialized.
func NewHandler(svc *service.Service, logger *logger.Logger) *Handler {
	return &Handler{
		Task:   NewTaskHandler(svc.Task, logger),
		logger: logger,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		tasks := api.Group("/tasks")
		{
			tasks.POST("", h.Task.Create)
			tasks.GET("", h.Task.List)
			tasks.GET("/:task_id", h.Task.Get)
			tasks.PUT("/:task_id", h.Task.Update)
			tasks.DELETE("/:task_id", h.Task.Delete)
		}
	}
}
