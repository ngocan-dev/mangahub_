package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient is a placeholder for MangaHub gRPC operations.
type GRPCClient struct {
	Address string
	conn    *grpc.ClientConn
}

// NewGRPCClient creates a new gRPC client wrapper.
func NewGRPCClient(address string) *GRPCClient {
	return &GRPCClient{Address: address}
}

// Connect establishes a gRPC connection.
func (c *GRPCClient) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	c.conn = conn
	return nil
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConn returns the underlying gRPC connection.
func (c *GRPCClient) GetConn() *grpc.ClientConn {
	return c.conn
}
