package client

import (
	"fmt"
	"net"
	"time"
)

// TCPClient handles TCP interactions.
type TCPClient struct {
	Address string
	Timeout time.Duration
}

// NewTCPClient constructs a TCP client wrapper.
func NewTCPClient(address string) *TCPClient {
	return &TCPClient{
		Address: address,
		Timeout: 10 * time.Second,
	}
}

// Dial dials the configured TCP address.
func (c *TCPClient) Dial() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", c.Address, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP server: %w", err)
	}
	return conn, nil
}

// SetTimeout sets the dial timeout.
func (c *TCPClient) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}
