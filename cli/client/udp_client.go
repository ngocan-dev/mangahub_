package client

import (
	"fmt"
	"net"
)

// UDPClient handles UDP interactions.
type UDPClient struct {
	Address string
}

// NewUDPClient constructs a UDP client wrapper.
func NewUDPClient(address string) *UDPClient {
	return &UDPClient{Address: address}
}

// Dial establishes a UDP connection.
func (c *UDPClient) Dial() (*net.UDPConn, error) {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Dial UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP server: %w", err)
	}

	return conn, nil
}

// Listen creates a UDP listener on the specified address.
func (c *UDPClient) Listen() (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP: %w", err)
	}

	return conn, nil
}
