# Example 04: Real-time WebSocket Server

Demonstrates real-time communication with WebSocket, connection hub management, and broadcasting patterns used in
production applications.

## Features

- **WebSocket Server**: Full-duplex real-time communication
- **Connection Hub**: Centralized connection management
- **Room/Channel System**: Group messaging
- **Broadcasting**: One-to-many message delivery
- **Connection Lifecycle**: Connect, disconnect, heartbeat
- **Message Types**: Join, leave, message, broadcast
- **Concurrent Safety**: Thread-safe connection map

## Architecture

```text
┌──────────────┐       WebSocket        ┌──────────────┐
│   Client 1   │◄─────────────────────►│              │
└──────────────┘                        │              │
┌──────────────┐       WebSocket        │     Hub      │
│   Client 2   │◄─────────────────────►│  (Gorilla)   │
└──────────────┘                        │              │
┌──────────────┐       WebSocket        │              │
│   Client 3   │◄─────────────────────►│              │
└──────────────┘                        └──────────────┘
                                              ▲
                                              │
                                        ┌─────┴──────┐
                                        │  Broadcast │
                                        │   System   │
                                        └────────────┘
```

## Features Demonstrated

### 1. Connection Hub

```go
type Hub struct {
    clients    map[*Client]bool
    rooms      map[string]map[*Client]bool
    broadcast  chan *Message
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
        case message := <-h.broadcast:
            h.broadcastToRoom(message)
        }
    }
}
```

### 2. Client Management

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    userID   string
    rooms    map[string]bool
}

func (c *Client) readPump() {
    // Read messages from WebSocket
}

func (c *Client) writePump() {
    // Write messages to WebSocket
}
```

### 3. Message Protocol

```json
{
  "type": "message",
  "room": "room1",
  "from": "user123",
  "content": "Hello, World!",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Project Structure

```text
04-realtime-websocket/
├── websocket/
│   ├── hub.go           # Connection hub
│   ├── client.go        # Client connection
│   └── handler.go       # WebSocket handler
├── main.go
├── web/
│   └── index.html       # WebSocket client demo
├── config.yaml
└── README.md
```

## Running

```bash
# Start server
go run main.go

# Open browser
open http://localhost:8080
```

## WebSocket Client Example

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = () => {
  // Join a room
  ws.send(
    JSON.stringify({
      type: "join",
      room: "general",
    })
  );
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};

// Send message
ws.send(
  JSON.stringify({
    type: "message",
    room: "general",
    content: "Hello!",
  })
);
```

## Message Types

| Type        | Description          | Example                                              |
|-------------|----------------------|------------------------------------------------------|
| `join`      | Join a room          | `{"type":"join","room":"general"}`                   |
| `leave`     | Leave a room         | `{"type":"leave","room":"general"}`                  |
| `message`   | Send message to room | `{"type":"message","room":"general","content":"Hi"}` |
| `broadcast` | Broadcast to all     | `{"type":"broadcast","content":"Announcement"}`      |
| `ping`      | Heartbeat            | `{"type":"ping"}`                                    |

- Chat applications
- Real-time dashboards
- Live notifications
- Collaborative editing
- Game servers
- Live feeds

## Testing

```bash
# Run tests
go test ./...
```

## Scaling Considerations

- **Redis Pub/Sub**: For multi-instance deployments
- **Message Queue**: For reliable delivery
- **Load Balancing**: Sticky sessions required
- **Connection Limits**: Monitor and limit per instance

## Next Steps

- Integrate with [authentication](../07-authentication)
- Add [event persistence](../06-event-driven)
- Scale with [Redis backend](../08-full-application)

## License

This example is part of the NCore project.
