package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// MessageType defines message types.
type MessageType string

const (
	MessageTypeJoin      MessageType = "join"
	MessageTypeLeave     MessageType = "leave"
	MessageTypeMessage   MessageType = "message"
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypePing      MessageType = "ping"
	MessageTypePong      MessageType = "pong"
)

// Message represents a WebSocket message.
type Message struct {
	Type      MessageType    `json:"type"`
	Room      string         `json:"room,omitempty"`
	From      string         `json:"from,omitempty"`
	Content   string         `json:"content,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Hub maintains active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *logger.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub.
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info(context.Background(), "Client registered", "client_id", client.id)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Remove from all rooms
				for room, clients := range h.rooms {
					if clients[client] {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.rooms, room)
						}
					}
				}
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info(context.Background(), "Client unregistered", "client_id", client.id)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			// Cleanup stale connections
			h.mu.RLock()
			count := len(h.clients)
			roomCount := len(h.rooms)
			h.mu.RUnlock()
			h.logger.Debug(context.Background(), "Hub stats", "clients", count, "rooms", roomCount)
		}
	}
}

// broadcastMessage sends a message to the appropriate clients.
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error(context.Background(), "Failed to marshal message", "error", err)
		return
	}

	if message.Room != "" {
		// Broadcast to room
		if clients, ok := h.rooms[message.Room]; ok {
			for client := range clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full, skip
					h.logger.Warn(context.Background(), "Client send buffer full", "client_id", client.id)
				}
			}
		}
	} else {
		// Broadcast to all
		for client := range h.clients {
			select {
			case client.send <- data:
			default:
				h.logger.Warn(context.Background(), "Client send buffer full", "client_id", client.id)
			}
		}
	}
}

// JoinRoom adds a client to a room.
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.rooms[room] = true

	h.logger.Info(context.Background(), "Client joined room", "client_id", client.id, "room", room)
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(client.rooms, room)

	h.logger.Info(context.Background(), "Client left room", "client_id", client.id, "room", room)
}

// GetStats returns hub statistics.
func (h *Hub) GetStats() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomSizes := make(map[string]int)
	for room, clients := range h.rooms {
		roomSizes[room] = len(clients)
	}

	return map[string]any{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"rooms":         roomSizes,
	}
}
