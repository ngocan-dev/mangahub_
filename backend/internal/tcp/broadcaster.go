package tcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/internal/queue"
)

// ServerBroadcaster implements the Broadcaster interface for the manga service
type ServerBroadcaster struct {
	server *Server
	queue  *queue.WriteQueue
}

// NewServerBroadcaster creates a new broadcaster that uses the TCP server
func NewServerBroadcaster(server *Server, queue *queue.WriteQueue) *ServerBroadcaster {
	return &ServerBroadcaster{server: server, queue: queue}
}

// BroadcastProgress broadcasts a progress update via TCP server
func (b *ServerBroadcaster) BroadcastProgress(ctx context.Context, userID, novelID int64, chapter int, chapterID *int64) error {
	var broadcastErr error

	if b.server != nil {
		if err := b.server.BroadcastProgress(ctx, userID, novelID, chapter, chapterID); err == nil {
			return nil
		} else {
			broadcastErr = err
		}
	} else {
		broadcastErr = errors.New("tcp server not configured")
	}

	if b.queue != nil {
		data := map[string]interface{}{
			"current_chapter": chapter,
		}
		if chapterID != nil {
			data["chapter_id"] = *chapterID
		}

		if err := b.queue.Enqueue("broadcast_progress", userID, novelID, data); err != nil {
			return fmt.Errorf("failed to broadcast and queue update: %w", err)
		}
	}

	return broadcastErr
}
