// Package handler provides HTTP handlers for the MongoDB example API.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/02-mongodb-api/service"
	"github.com/ncobase/ncore/logging/logger"
)

// Handler aggregates all HTTP handlers.
type Handler struct {
	User   *UserHandler
	logger *logger.Logger
}

// NewHandler creates a new handler instance with all sub-handlers initialized.
func NewHandler(svc *service.Service, logger *logger.Logger) *Handler {
	return &Handler{
		User:   NewUserHandler(svc.User, logger),
		logger: logger,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", h.User.Create)
			users.GET("", h.User.List)
			users.GET("/:user_id", h.User.Get)
			users.PUT("/:user_id", h.User.Update)
			users.DELETE("/:user_id", h.User.Delete)
		}
	}
}
