package tcp

import (
	"net"
	"sync"
	"time"
)

// Client represents a connected TCP client
type Client struct {
	Conn          net.Conn
	UserID        int64
	Username      string
	DeviceName    string
	DeviceType    string
	ConnectedAt   time.Time
	LastSeen      time.Time
	mu            sync.RWMutex
	authenticated bool
}

// NewClient creates a new client instance
func NewClient(conn net.Conn) *Client {
	now := time.Now()
	return &Client{
		Conn:          conn,
		ConnectedAt:   now,
		LastSeen:      now,
		authenticated: false,
	}
}

// IsAuthenticated returns whether the client is authenticated
func (c *Client) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// SetAuthenticated sets the authentication status
func (c *Client) SetAuthenticated(userID int64, username, deviceName, deviceType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authenticated = true
	c.UserID = userID
	c.Username = username
	c.DeviceName = deviceName
	c.DeviceType = deviceType
	c.LastSeen = time.Now()
}

// UpdateLastSeen updates the last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastSeen = time.Now()
}

// SendMessage sends a message to the client
// A2: Send fails - Server logs error and continues with other clients
func (c *Client) SendMessage(msg *Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := SerializeMessage(msg)
	if err != nil {
		return err
	}

	// Add newline delimiter
	data = append(data, '\n')

	// Set write deadline
	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = c.Conn.Write(data)
	if err != nil {
		// A1: Connection lost - return error so server can remove client
		return err
	}

	c.LastSeen = time.Now()
	return nil
}

// SendError sends an error message to the client
func (c *Client) SendError(code, message string) error {
	msg := &Message{
		Type:  MessageTypeError,
		Error: message,
		Payload: ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
	return c.SendMessage(msg)
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.Close()
}

// ReadMessage reads a message from the client connection
func (c *Client) ReadMessage() (*Message, error) {
	// Set read deadline
	c.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Read until newline
	var buffer []byte
	buf := make([]byte, 4096)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			return nil, err
		}

		buffer = append(buffer, buf[:n]...)

		// Check for newline delimiter
		for i, b := range buffer {
			if b == '\n' {
				msg, err := ParseMessage(buffer[:i])
				if err != nil {
					return nil, err
				}
				c.UpdateLastSeen()
				return msg, nil
			}
		}

		// Prevent buffer overflow
		if len(buffer) > 65536 {
			return nil, ErrInvalidMessage
		}
	}
}
