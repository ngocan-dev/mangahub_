package client

import "net"

// TCPClient handles placeholder TCP interactions.
type TCPClient struct {
	Address string
}

// NewTCPClient constructs a TCP client wrapper.
func NewTCPClient(address string) *TCPClient {
	return &TCPClient{Address: address}
}

// Dial dials the configured TCP address.
func (c *TCPClient) Dial() (net.Conn, error) {
	// TODO: implement TCP dial logic
	return nil, nil
}
