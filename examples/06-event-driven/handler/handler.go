// Package handler exposes HTTP endpoints for event-driven workflows.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/06-event-driven/event"
	"github.com/ncobase/ncore/examples/06-event-driven/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// Handler aggregates all HTTP handlers.
type Handler struct {
	userService *service.UserService
	eventBus    *event.Bus
	eventStore  event.EventStore
	analytics   *service.AnalyticsService
	logger      *logger.Logger
}

// NewHandler creates a new handler.
func NewHandler(
	userService *service.UserService,
	eventBus *event.Bus,
	eventStore event.EventStore,
	analytics *service.AnalyticsService,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		userService: userService,
		eventBus:    eventBus,
		eventStore:  eventStore,
		analytics:   analytics,
		logger:      logger,
	}
}

// RegisterUser handles user registration.
func (h *Handler) RegisterUser(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to register user"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, user)
}

// UpdateUser handles user updates.
func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req.Name, req.Email)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("user not found"))
		return
	}

	resp.Success(c.Writer, user)
}

// GetUser retrieves a user.
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := h.userService.GetUser(userID)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("user not found"))
		return
	}

	resp.Success(c.Writer, user)
}

// ListUsers lists all users.
func (h *Handler) ListUsers(c *gin.Context) {
	users := h.userService.ListUsers()
	resp.Success(c.Writer, users)
}

// PublishEvent publishes a custom event.
func (h *Handler) PublishEvent(c *gin.Context) {
	var req struct {
		Type        string         `json:"type" binding:"required"`
		AggregateID string         `json:"aggregate_id"`
		Payload     map[string]any `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	evt := &event.Event{
		Type:        event.EventType(req.Type),
		AggregateID: req.AggregateID,
		Payload:     req.Payload,
	}

	if err := h.eventBus.Publish(c.Request.Context(), evt); err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to publish event"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusAccepted, evt)
}

// GetEvents retrieves events.
func (h *Handler) GetEvents(c *gin.Context) {
	eventType := c.Query("type")
	aggregateID := c.Query("aggregate_id")

	var events []*event.Event
	var err error

	if aggregateID != "" {
		events, err = h.eventStore.LoadByAggregate(c.Request.Context(), aggregateID)
	} else if eventType != "" {
		events, err = h.eventStore.LoadByType(c.Request.Context(), event.EventType(eventType))
	}

	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to load events"))
		return
	}

	resp.Success(c.Writer, events)
}

// GetStats returns event bus and analytics statistics.
func (h *Handler) GetStats(c *gin.Context) {
	stats := map[string]any{
		"event_bus": h.eventBus.GetStats(),
		"analytics": h.analytics.GetStats(),
	}

	resp.Success(c.Writer, stats)
}
