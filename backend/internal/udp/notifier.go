package udp

import (
	"context"
	"log"
	"net"
	"time"
)

// Notifier handles sending notifications to registered clients
type Notifier struct {
	server *Server
}

// NewNotifier creates a new notifier instance
func NewNotifier(server *Server) *Notifier {
	return &Notifier{server: server}
}

// NotifyChapterRelease sends a chapter release notification to subscribed clients
func (n *Notifier) NotifyChapterRelease(ctx context.Context, novelID int64, novelName string, chapter int, chapterID int64) error {
	if n.server == nil {
		return nil
	}

	notification := &Packet{
		Type: PacketTypeNotification,
		Payload: NotificationPacket{
			NovelID:   novelID,
			NovelName: novelName,
			Chapter:   chapter,
			ChapterID: chapterID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Get clients subscribed to this novel
	n.server.mu.RLock()
	clientMap := make(map[string]*Client) // Use map to avoid duplicates

	// Get clients subscribed to all novels
	for _, clientList := range n.server.clientsByUser {
		for _, c := range clientList {
			if c.IsSubscribedTo(novelID) {
				clientMap[c.GetKey()] = c
			}
		}
	}

	// Get clients subscribed to this specific novel
	if novelClients, ok := n.server.clientsByNovel[novelID]; ok {
		for _, c := range novelClients {
			clientMap[c.GetKey()] = c
		}
	}

	// Convert map to slice
	clients := make([]*Client, 0, len(clientMap))
	for _, c := range clientMap {
		clients = append(clients, c)
	}

	n.server.mu.RUnlock()

	// Step 3: Broadcast message to all registered clients
	// Main Success Scenario:
	// 1. Administrator triggers notification for specific manga
	// 2. System creates notification message with manga details
	// 3. UDP server broadcasts message to all registered clients
	// 4. Clients receive notification and display to users
	// 5. System logs successful broadcast
	successCount := 0
	failedCount := 0

	for _, client := range clients {
		client.UpdateLastSeen()

		// A2: Network error - Server logs error and retries
		err := n.sendWithRetry(ctx, client.Address, notification, 3)
		if err != nil {
			// A1: Client unreachable - Server continues with other clients
			log.Printf("Failed to send notification to %s (UserID=%d) after retries: %v",
				client.Address.String(), client.UserID, err)
			failedCount++
			continue
		}
		successCount++
	}

	// Step 5: System logs successful broadcast
	log.Printf("Chapter release notification broadcasted: NovelID=%d, NovelName=%s, Chapter=%d, Sent to %d/%d clients (Failed: %d)",
		novelID, novelName, chapter, successCount, len(clients), failedCount)

	return nil
}

// sendWithRetry sends a packet with retry logic
// A2: Network error - Server logs error and retries
func (n *Notifier) sendWithRetry(ctx context.Context, addr *net.UDPAddr, packet *Packet, maxRetries int) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := n.server.sendPacket(addr, packet)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a network error that might be retryable
		if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
			if attempt < maxRetries {
				// Wait before retry (exponential backoff)
				waitTime := time.Duration(attempt) * 100 * time.Millisecond
				log.Printf("Network error sending to %s (attempt %d/%d), retrying in %v: %v",
					addr.String(), attempt, maxRetries, waitTime, err)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(waitTime):
					continue
				}
			}
		}

		// Non-retryable error or max retries reached
		break
	}

	return lastErr
}
