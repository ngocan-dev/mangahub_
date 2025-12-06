package client

import "context"

// GRPCClient is a placeholder for MangaHub gRPC operations.
type GRPCClient struct {
	Address string
}

// NewGRPCClient creates a new gRPC client wrapper.
func NewGRPCClient(address string) *GRPCClient {
	return &GRPCClient{Address: address}
}

// Connect establishes a placeholder connection.
func (c *GRPCClient) Connect(ctx context.Context) error {
	// TODO: implement gRPC connection logic
	return nil
}
