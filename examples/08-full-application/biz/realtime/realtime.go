// Package realtime provides WebSocket collaboration for the full app.
package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"

	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a WebSocket client.
type Client struct {
	ID          string
	WorkspaceID string
	UserID      string
	Conn        *websocket.Conn
	Send        chan []byte
	hub         *Hub
}

// Hub manages WebSocket clients and broadcasts.
type Hub struct {
	// Registered clients by workspace
	clients    map[string]map[*Client]bool // workspaceID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
	logger     *logger.Logger
}

// Message represents a WebSocket message.
type Message struct {
	Type      string         `json:"type"`
	Workspace string         `json:"workspace_id,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	Data      map[string]any `json:"data"`
	Timestamp int64          `json:"timestamp"`
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		logger:     logger,
	}
}

// Run starts the hub.
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info(context.Background(), "WebSocket hub starting")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info(context.Background(), "WebSocket hub stopping")
			return
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client.
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.WorkspaceID]; !exists {
		h.clients[client.WorkspaceID] = make(map[*Client]bool)
	}

	h.clients[client.WorkspaceID][client] = true
	h.logger.Info(context.Background(), "WebSocket client registered",
		"client_id", client.ID,
		"workspace_id", client.WorkspaceID,
		"user_id", client.UserID,
		"total_clients", h.countClients())
}

// unregisterClient unregisters a client.
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.clients[client.WorkspaceID]; exists {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			h.logger.Info(context.Background(), "WebSocket client unregistered",
				"client_id", client.ID,
				"workspace_id", client.WorkspaceID,
				"total_clients", h.countClients())

			// Remove workspace if no clients left
			if len(clients) == 0 {
				delete(h.clients, client.WorkspaceID)
			}
		}
	}
}

// broadcastMessage broadcasts a message to all clients in a workspace.
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message.Timestamp = time.Now().Unix()
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error(context.Background(), "Failed to marshal WebSocket message", "error", err)
		return
	}

	// Broadcast to all clients in workspace or to all workspaces
	if message.Workspace != "" {
		if clients, exists := h.clients[message.Workspace]; exists {
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					h.unregisterClient(client)
				}
			}
		}
	} else {
		// Broadcast to all clients
		for _, clients := range h.clients {
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					h.unregisterClient(client)
				}
			}
		}
	}
}

// countClients returns the total number of connected clients.
func (h *Hub) countClients() int {
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// GetStats returns hub statistics.
func (h *Hub) GetStats() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]any)
	for workspaceID, clients := range h.clients {
		stats[workspaceID] = len(clients)
	}

	return map[string]any{
		"total_clients":    h.countClients(),
		"total_workspaces": len(h.clients),
		"workspace_stats":  stats,
	}
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error(context.Background(), "WebSocket read error", "error", err, "client_id", c.ID)
			}
			break
		}

		// Parse message and handle
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.logger.Error(context.Background(), "Failed to unmarshal WebSocket message", "error", err, "client_id", c.ID)
			continue
		}

		// Broadcast to workspace
		msg.Workspace = c.WorkspaceID
		msg.UserID = c.UserID
		c.hub.broadcast <- &msg
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.hub.logger.Error(context.Background(), "WebSocket write error", "error", err, "client_id", c.ID)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type Module struct {
	types.OptionalImpl

	hub    *Hub
	bus    *event.Bus
	logger *logger.Logger
}

func init() {
	registry.RegisterToGroup(New(), "biz")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "realtime"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"auth"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Realtime WebSocket module",
		Type:         "module",
		Group:        "biz",
		Dependencies: m.Dependencies(),
	}
}

func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	busAny, err := em.GetCrossService("app", "EventBus")
	if err != nil {
		return err
	}
	bus, ok := busAny.(*event.Bus)
	if !ok {
		return fmt.Errorf("event bus type mismatch")
	}
	m.bus = bus

	m.hub = NewHub(m.logger)
	go m.hub.Run(context.Background())

	m.bus.Subscribe(event.EventTypeTaskCreated, m.handleTaskEvent)
	m.bus.Subscribe(event.EventTypeTaskUpdated, m.handleTaskEvent)
	m.bus.Subscribe(event.EventTypeTaskDeleted, m.handleTaskEvent)
	m.bus.Subscribe(event.EventTypeTaskAssigned, m.handleTaskEvent)
	m.bus.Subscribe(event.EventTypeCommentCreated, m.handleCommentEvent)
	m.bus.Subscribe(event.EventTypeCommentUpdated, m.handleCommentEvent)
	m.bus.Subscribe(event.EventTypeCommentDeleted, m.handleCommentEvent)

	m.logger.Info(context.Background(), "Realtime module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Realtime module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return nil
}

func (m *Module) GetServices() types.Service {
	return m.hub
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/ws/:workspace_id", m.HandleWebSocket)
}

func (m *Module) HandleWebSocket(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID, exists := c.Get("user_id")
	if !exists {
		resp.Fail(c.Writer, resp.UnAuthorized("not authenticated"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error(c.Request.Context(), "Failed to upgrade WebSocket connection", "error", err)
		return
	}

	client := &Client{
		ID:          fmt.Sprintf("%s-%s", workspaceID, userID.(string)),
		WorkspaceID: workspaceID,
		UserID:      userID.(string),
		Conn:        conn,
		Send:        make(chan []byte, 256),
		hub:         m.hub,
	}

	m.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "Realtime module cleanup", "module", m.Name())
	return nil
}

func (m *Module) Hub() *Hub {
	return m.hub
}

func (m *Module) handleTaskEvent(ctx context.Context, evt *event.Event) error {
	message := &Message{
		Type:      string(evt.Type),
		Workspace: evt.WorkspaceID,
		UserID:    evt.UserID,
		Data:      evt.Payload,
	}

	m.hub.broadcast <- message
	return nil
}

func (m *Module) handleCommentEvent(ctx context.Context, evt *event.Event) error {
	message := &Message{
		Type:      string(evt.Type),
		Workspace: evt.WorkspaceID,
		UserID:    evt.UserID,
		Data:      evt.Payload,
	}

	m.hub.broadcast <- message
	return nil
}
