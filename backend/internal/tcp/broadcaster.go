package tcp

import (
	"context"
)

// ServerBroadcaster implements the Broadcaster interface for the manga service
type ServerBroadcaster struct {
	server *Server
}

// NewServerBroadcaster creates a new broadcaster that uses the TCP server
func NewServerBroadcaster(server *Server) *ServerBroadcaster {
	return &ServerBroadcaster{server: server}
}

// BroadcastProgress broadcasts a progress update via TCP server
func (b *ServerBroadcaster) BroadcastProgress(ctx context.Context, userID, novelID int64, chapter int, chapterID *int64) error {
	if b.server == nil {
		return nil // Server not available
	}
	return b.server.BroadcastProgress(ctx, userID, novelID, chapter, chapterID)
}
