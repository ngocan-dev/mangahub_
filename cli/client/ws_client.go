package client

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient handles WebSocket interactions.
type WSClient struct {
	URL    string
	Dial   *url.URL
	conn   *websocket.Conn
	dialer *websocket.Dialer
}

// NewWSClient constructs a WebSocket client wrapper.
func NewWSClient(urlStr string) *WSClient {
	u, _ := url.Parse(urlStr)
	return &WSClient{
		URL:  urlStr,
		Dial: u,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Connect establishes a WebSocket connection.
func (c *WSClient) Connect() error {
	conn, _, err := c.dialer.Dial(c.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	c.conn = conn
	return nil
}

// Close closes the WebSocket connection.
func (c *WSClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConn returns the underlying WebSocket connection.
func (c *WSClient) GetConn() *websocket.Conn {
	return c.conn
}

// SetTimeout sets the handshake timeout.
func (c *WSClient) SetTimeout(timeout time.Duration) {
	c.dialer.HandshakeTimeout = timeout
}
