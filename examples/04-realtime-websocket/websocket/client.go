package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ncobase/ncore/logging/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client.
type Client struct {
	id     string
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	rooms  map[string]bool
	logger *logger.Logger
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, logger *logger.Logger) *Client {
	return &Client{
		id:     uuid.New().String(),
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		rooms:  make(map[string]bool),
		logger: logger,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error(context.Background(), "WebSocket read error", "client_id", c.id, "error", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Warn(context.Background(), "Invalid message format", "client_id", c.id, "error", err)
			continue
		}

		msg.From = c.id
		msg.Timestamp = time.Now()

		c.handleMessage(&msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Error(context.Background(), "WebSocket write error", "client_id", c.id, "error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages.
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeJoin:
		if msg.Room != "" {
			c.hub.JoinRoom(c, msg.Room)
		}

	case MessageTypeLeave:
		if msg.Room != "" {
			c.hub.LeaveRoom(c, msg.Room)
		}

	case MessageTypeMessage:
		c.hub.broadcast <- msg

	case MessageTypeBroadcast:
		msg.Room = "" // Broadcast to all
		c.hub.broadcast <- msg

	case MessageTypePing:
		// Send pong
		pong := &Message{
			Type:      MessageTypePong,
			Timestamp: time.Now(),
		}
		data, _ := json.Marshal(pong)
		select {
		case c.send <- data:
		default:
		}
	}
}
