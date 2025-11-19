package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	UserID   int64
	Username string
	RoomID   int64
	mu       sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// SetUser sets the user information
func (c *Client) SetUser(userID int64, username string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.UserID = userID
	c.Username = username
}

// SetRoom sets the room ID
func (c *Client) SetRoom(roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RoomID = roomID
}

// GetUserID returns the user ID
func (c *Client) GetUserID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.UserID
}

// GetUsername returns the username
func (c *Client) GetUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Username
}

// GetRoomID returns the room ID
func (c *Client) GetRoomID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RoomID
}

// ReadPump pumps messages from the WebSocket connection to the hub
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		msg, err := ParseMessage(message)
		if err != nil {
			log.Printf("Error parsing message: %v", err)
			c.SendError("invalid_message", "invalid message format")
			continue
		}

		// Handle message
		c.hub.handleMessage(c, msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg *Message) error {
	data, err := SerializeMessage(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
	default:
		// Channel full - client is too slow
		log.Printf("Client send buffer full, dropping message")
	}
	return nil
}

// SendError sends an error message to the client
func (c *Client) SendError(code, message string) {
	msg := &Message{
		Type:  MessageTypeError,
		Error: message,
		Payload: map[string]string{
			"code":    code,
			"message": message,
		},
	}
	c.SendMessage(msg)
}
