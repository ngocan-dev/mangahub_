package tcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/auth"
)

// Server represents the TCP server
type Server struct {
	address       string
	maxClients    int
	db            *sql.DB
	clients       map[*Client]bool
	clientsByUser map[int64][]*Client // Multiple devices per user
	mu            sync.RWMutex
	broadcastCh   chan ProgressUpdate
}

// NewServer creates a new TCP server instance
// TCP and WebSocket connections remain stable
func NewServer(address string, maxClients int, db *sql.DB) *Server {
	return &Server{
		address:       address,
		maxClients:    maxClients,
		db:            db,
		clients:       make(map[*Client]bool),
		clientsByUser: make(map[int64][]*Client),
		broadcastCh:   make(chan ProgressUpdate, 1000), // Increased buffer for 50-100 concurrent users
	}
}

// Start starts the TCP server
func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("TCP server listening on %s", s.address)

	// Start broadcast handler
	go s.handleBroadcasts(ctx)

	// Accept connections
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			// Set connection timeouts for stability
			// TCP and WebSocket connections remain stable
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// A2: Check server capacity
			s.mu.RLock()
			currentClients := len(s.clients)
			s.mu.RUnlock()

			if currentClients >= s.maxClients {
				log.Printf("Server at capacity (%d/%d), rejecting connection", currentClients, s.maxClients)
				conn.Close()
				continue
			}

			// Step 2: Create goroutine handler for each connection
			client := NewClient(conn)
			go s.handleClient(ctx, client)
		}
	}
}

// handleClient handles a client connection
func (s *Server) handleClient(ctx context.Context, client *Client) {
	defer func() {
		s.removeClient(client)
		client.Close()
	}()

	log.Printf("New client connected from %s", client.Conn.RemoteAddr())

	// Step 3: Wait for authentication message
	authTimeout := time.NewTimer(30 * time.Second)
	defer authTimeout.Stop()

	authChan := make(chan bool, 1)
	doneChan := make(chan bool, 1)

	go func() {
		defer func() { doneChan <- true }()
		for {
			select {
			case <-ctx.Done():
				return
			case <-authTimeout.C:
				// Authentication timeout
				client.SendError("auth_timeout", "authentication timeout")
				authChan <- false
				return
			default:
				msg, err := client.ReadMessage()
				if err != nil {
					log.Printf("Error reading message from client: %v", err)
					authChan <- false
					return
				}

				if msg.Type == MessageTypeAuth {
					authenticated := s.handleAuthentication(client, msg)
					authChan <- authenticated
					if !authenticated {
						return // A1: Authentication fails - Server closes connection
					}
					authTimeout.Stop()
					return
				} else {
					// Client must authenticate first
					client.SendError("auth_required", "authentication required")
				}
			}
		}
	}()

	// Wait for authentication
	select {
	case <-ctx.Done():
		return
	case authenticated := <-authChan:
		if !authenticated {
			<-doneChan // Wait for goroutine to finish
			return
		}
		<-doneChan // Wait for goroutine to finish
	}

	// Step 5: Client is authenticated, handle messages
	log.Printf("Client authenticated: UserID=%d, Username=%s", client.UserID, client.Username)

	// Heartbeat ticker
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			// Send heartbeat
			msg := &Message{
				Type: MessageTypeHeartbeat,
			}
			if err := client.SendMessage(msg); err != nil {
				// A1: Client connection lost - Server removes from active list
				log.Printf("Error sending heartbeat to client (UserID=%d): %v", client.UserID, err)
				return
			}
		default:
			// Set read timeout
			client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			msg, err := client.ReadMessage()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout - continue to heartbeat
					continue
				}
				// A1: Client connection lost - Server removes from active list
				log.Printf("Error reading message from client (UserID=%d): %v", client.UserID, err)
				return
			}

			// Handle different message types
			switch msg.Type {
			case MessageTypeHeartbeat:
				// Acknowledge heartbeat
				client.UpdateLastSeen()
			default:
				log.Printf("Unknown message type: %s", msg.Type)
			}
		}
	}
}

// handleAuthentication handles client authentication
func (s *Server) handleAuthentication(client *Client, msg *Message) bool {
	// Parse auth request
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		client.SendError("invalid_request", "invalid authentication request")
		return false
	}

	var authReq AuthRequest
	if err := json.Unmarshal(payloadBytes, &authReq); err != nil {
		client.SendError("invalid_request", "invalid authentication request format")
		return false
	}

	// Step 4: Validate JWT token
	// Invalid tokens are rejected
	// Expired tokens trigger reauthentication
	claims, err := auth.ValidateToken(authReq.Token)
	if err != nil {
		// Handle different error types
		if errors.Is(err, auth.ErrExpiredToken) {
			// Expired tokens trigger reauthentication
			client.SendError("token_expired", "your session has expired. please login again")
		} else if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrInvalidSigningMethod) {
			// Invalid tokens are rejected
			client.SendError("auth_failed", "invalid token")
		} else if errors.Is(err, auth.ErrInvalidClaims) {
			// Token claims are properly validated
			client.SendError("auth_failed", "invalid token claims")
		} else {
			client.SendError("auth_failed", "authentication failed")
		}
		return false
	}
	if claims == nil {
		// Unauthorized access is prevented
		client.SendError("auth_failed", "token validation failed")
		return false
	}

	// Register client
	client.SetAuthenticated(claims.UserID, claims.Username, authReq.DeviceName, authReq.DeviceType)
	s.addClient(client)

	// Step 5: Send confirmation
	resp := &Message{
		Type: MessageTypeAuthResp,
		Payload: AuthResponse{
			Success:  true,
			UserID:   claims.UserID,
			Username: claims.Username,
			Message:  "authentication successful",
		},
	}

	if err := client.SendMessage(resp); err != nil {
		log.Printf("Error sending auth response: %v", err)
		return false
	}

	// Record sync session in database
	go s.recordSyncSession(client)

	return true
}

// addClient adds a client to the active list
func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = true
	s.clientsByUser[client.UserID] = append(s.clientsByUser[client.UserID], client)
	log.Printf("Client registered: UserID=%d, Total clients: %d", client.UserID, len(s.clients))
}

// removeClient removes a client from the active list
func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, client)

	// Remove from user's client list
	if clients, ok := s.clientsByUser[client.UserID]; ok {
		for i, c := range clients {
			if c == client {
				s.clientsByUser[client.UserID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(s.clientsByUser[client.UserID]) == 0 {
			delete(s.clientsByUser, client.UserID)
		}
	}

	log.Printf("Client disconnected: UserID=%d, Total clients: %d", client.UserID, len(s.clients))
}

// BroadcastProgress broadcasts a progress update to all clients
func (s *Server) BroadcastProgress(ctx context.Context, userID, novelID int64, chapter int, chapterID *int64) error {
	update := ProgressUpdate{
		UserID:    userID,
		NovelID:   novelID,
		Chapter:   chapter,
		ChapterID: chapterID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	select {
	case s.broadcastCh <- update:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel full - log but don't block
		log.Printf("Broadcast channel full, dropping update")
		return nil
	}
}

// handleBroadcasts handles broadcasting progress updates
// Main Success Scenario:
// 1. System receives progress update from HTTP API
// 2. TCP server receives broadcast message via channel
// 3. Server identifies connections for the specific user
// 4. Server sends JSON progress message to connections
// 5. Clients receive and process update
func (s *Server) handleBroadcasts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-s.broadcastCh:
			// Step 3: Identify connections for the specific user
			s.mu.RLock()
			userClients, exists := s.clientsByUser[update.UserID]
			if !exists || len(userClients) == 0 {
				s.mu.RUnlock()
				log.Printf("No active connections for user %d", update.UserID)
				continue
			}

			// Create a copy of the client list to avoid holding lock during sends
			clients := make([]*Client, len(userClients))
			copy(clients, userClients)
			s.mu.RUnlock()

			// Step 4: Send JSON progress message to connections
			msg := &Message{
				Type:    MessageTypeProgress,
				Payload: update,
			}

			successCount := 0
			for _, client := range clients {
				// Check if client is still authenticated
				if !client.IsAuthenticated() {
					// A1: Client connection lost - Server removes from active list
					s.removeClient(client)
					continue
				}

				// Send message to client
				if err := client.SendMessage(msg); err != nil {
					// A2: Send fails - Server logs error and continues with other clients
					log.Printf("Error broadcasting to client (UserID=%d, Device=%s): %v",
						client.UserID, client.DeviceName, err)
					// A1: Remove failed client from active list
					s.removeClient(client)
					continue
				}

				successCount++
			}

			log.Printf("Progress update broadcasted: UserID=%d, NovelID=%d, Chapter=%d, Sent to %d/%d clients",
				update.UserID, update.NovelID, update.Chapter, successCount, len(clients))
		}
	}
}

// recordSyncSession records the sync session in the database
func (s *Server) recordSyncSession(client *Client) {
	if s.db == nil {
		return
	}

	_, err := s.db.Exec(`
		INSERT INTO Sync_Sessions (User_Id, Device_Name, Device_Type, Status, Started_At, Last_Seen_At, Last_Ip)
		VALUES (?, ?, ?, 'active', ?, ?, ?)
	`, client.UserID, client.DeviceName, client.DeviceType, client.ConnectedAt, client.LastSeen, client.Conn.RemoteAddr().String())

	if err != nil {
		log.Printf("Error recording sync session: %v", err)
	}
}

// GetClientCount returns the current number of connected clients
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}
