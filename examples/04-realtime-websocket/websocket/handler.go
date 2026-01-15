// Package websocket implements the realtime hub and handlers.
package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Handler handles WebSocket connections.
type Handler struct {
	hub    *Hub
	logger *logger.Logger
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, logger *logger.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// HandleConnection handles WebSocket upgrade and connection.
func (h *Handler) HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Failed to upgrade connection", "error", err)
		return
	}

	client := NewClient(h.hub, conn, h.logger)
	h.hub.register <- client

	// Start goroutines
	go client.WritePump()
	go client.ReadPump()
}

// HandleStats returns WebSocket statistics.
func (h *Handler) HandleStats(c *gin.Context) {
	stats := h.hub.GetStats()
	resp.Success(c.Writer, stats)
}
