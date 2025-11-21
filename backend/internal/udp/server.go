package udp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ngocan-dev/mangahub/backend/internal/auth"
)

// Server represents the UDP notification server
type Server struct {
	address        string
	conn           *net.UDPConn
	db             *sql.DB
	clients        map[string]*Client  // Key: address:userID
	clientsByUser  map[int64][]*Client // Clients grouped by user
	clientsByNovel map[int64][]*Client // Clients grouped by novel subscription
	mu             sync.RWMutex
}

// NewServer creates a new UDP server instance
func NewServer(address string, db *sql.DB) *Server {
	return &Server{
		address:        address,
		db:             db,
		clients:        make(map[string]*Client),
		clientsByUser:  make(map[int64][]*Client),
		clientsByNovel: make(map[int64][]*Client),
	}
}

// Start starts the UDP server
func (s *Server) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	s.conn = conn
	defer conn.Close()

	log.Printf("UDP notification server listening on %s", s.address)

	// Start cleanup goroutine for stale clients
	go s.cleanupStaleClients(ctx)

	// Main receive loop
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Set read deadline
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout - continue loop
					continue
				}
				log.Printf("Error reading UDP packet: %v", err)
				continue
			}

			// Step 2: Server receives registration and extracts client address
			go s.handlePacket(ctx, buffer[:n], clientAddr)
		}
	}
}

// handlePacket handles an incoming UDP packet
// Main Success Scenario:
// 1. Client sends UDP registration packet with user preferences
// 2. Server receives registration and extracts client address
// 3. Server adds client to notification list
// 4. Server sends confirmation packet to client
// 5. Client is ready to receive notifications
func (s *Server) handlePacket(ctx context.Context, data []byte, addr *net.UDPAddr) {
	packet, err := ParsePacket(data)
	if err != nil {
		log.Printf("Error parsing packet from %s: %v", addr.String(), err)
		s.sendError(addr, "invalid_packet", "invalid packet format")
		return
	}

	switch packet.Type {
	case PacketTypeRegister:
		s.handleRegister(ctx, packet, addr)
	case PacketTypeUnregister:
		s.handleUnregister(ctx, packet, addr)
	default:
		log.Printf("Unknown packet type: %s from %s", packet.Type, addr.String())
		s.sendError(addr, "unknown_type", "unknown packet type")
	}
}

// handleRegister handles client registration
func (s *Server) handleRegister(ctx context.Context, packet *Packet, addr *net.UDPAddr) {
	// Parse registration request
	payloadBytes, err := json.Marshal(packet.Payload)
	if err != nil {
		s.sendError(addr, "invalid_request", "invalid registration request")
		return
	}

	var req RegisterRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		s.sendError(addr, "invalid_request", "invalid registration request format")
		return
	}

	// Validate user ID or token
	userID := req.UserID
	if userID == 0 && req.Token != "" {
		// Validate JWT token
		claims, err := auth.ValidateToken(req.Token)
		if err != nil || claims == nil {
			s.sendError(addr, "auth_failed", "invalid or expired token")
			return
		}
		userID = claims.UserID
	}

	if userID == 0 {
		s.sendError(addr, "invalid_request", "user_id or token required")
		return
	}

	// Step 3: Add client to notification list
	clientKey := fmt.Sprintf("%s:%d", addr.String(), userID)

	// Check if already registered
	s.mu.RLock()
	existingClient, exists := s.clients[clientKey]
	s.mu.RUnlock()

	if exists {
		// Update existing registration
		existingClient.UpdateLastSeen()
		// Update novel subscriptions if provided
		if len(req.NovelIDs) > 0 || req.AllNovels {
			s.updateClientSubscriptions(existingClient, req.NovelIDs, req.AllNovels)
		}
	} else {
		// Create new client
		novelIDs := req.NovelIDs
		if req.AllNovels {
			novelIDs = []int64{} // Empty means all
		}

		client := NewClient(addr, userID, novelIDs, req.AllNovels, req.DeviceID)
		s.addClient(client)

		// Record subscription in database
		go s.recordSubscription(ctx, client)
	}

	// Step 4: Send confirmation packet to client
	resp := &Packet{
		Type: PacketTypeConfirm,
		Payload: RegisterResponse{
			Success:  true,
			UserID:   userID,
			Message:  "registration successful",
			NovelIDs: req.NovelIDs,
		},
	}

	if err := s.sendPacket(addr, resp); err != nil {
		log.Printf("Error sending confirmation to %s: %v", addr.String(), err)
		return
	}

	log.Printf("Client registered: UserID=%d, Address=%s, Novels=%v",
		userID, addr.String(), req.NovelIDs)
}

// handleUnregister handles client unregistration
func (s *Server) handleUnregister(ctx context.Context, packet *Packet, addr *net.UDPAddr) {
	payloadBytes, err := json.Marshal(packet.Payload)
	if err != nil {
		return
	}

	var req RegisterRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return
	}

	userID := req.UserID
	if userID == 0 && req.Token != "" {
		claims, err := auth.ValidateToken(req.Token)
		if err == nil && claims != nil {
			userID = claims.UserID
		}
	}

	if userID == 0 {
		return
	}

	clientKey := fmt.Sprintf("%s:%d", addr.String(), userID)
	s.removeClient(clientKey)

	log.Printf("Client unregistered: UserID=%d, Address=%s", userID, addr.String())
}

// addClient adds a client to the notification list
func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientKey := client.GetKey()
	s.clients[clientKey] = client

	// Add to user's client list
	s.clientsByUser[client.UserID] = append(s.clientsByUser[client.UserID], client)

	// Add to novel subscription lists
	if client.AllNovels {
		// If subscribed to all, we'll check this separately
	} else {
		for _, novelID := range client.NovelIDs {
			s.clientsByNovel[novelID] = append(s.clientsByNovel[novelID], client)
		}
	}
}

// updateClientSubscriptions updates a client's novel subscriptions
func (s *Server) updateClientSubscriptions(client *Client, novelIDs []int64, allNovels bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from old novel lists
	if !client.AllNovels {
		for _, novelID := range client.NovelIDs {
			clients := s.clientsByNovel[novelID]
			for i, c := range clients {
				if c == client {
					s.clientsByNovel[novelID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
		}
	}

	// Update client subscriptions
	client.mu.Lock()
	client.NovelIDs = novelIDs
	client.AllNovels = allNovels
	client.mu.Unlock()

	// Add to new novel lists
	if allNovels {
		// Subscribed to all - no need to add to specific lists
	} else {
		for _, novelID := range novelIDs {
			s.clientsByNovel[novelID] = append(s.clientsByNovel[novelID], client)
		}
	}
}

// removeClient removes a client from the notification list
func (s *Server) removeClient(clientKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientKey]
	if !exists {
		return
	}

	delete(s.clients, clientKey)

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

	// Remove from novel subscription lists
	if !client.AllNovels {
		for _, novelID := range client.NovelIDs {
			if clients, ok := s.clientsByNovel[novelID]; ok {
				for i, c := range clients {
					if c == client {
						s.clientsByNovel[novelID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
			}
		}
	}
}

// sendPacket sends a UDP packet to an address
func (s *Server) sendPacket(addr *net.UDPAddr, packet *Packet) error {
	data, err := SerializePacket(packet)
	if err != nil {
		return err
	}

	_, err = s.conn.WriteToUDP(data, addr)
	return err
}

// sendError sends an error packet to a client
func (s *Server) sendError(addr *net.UDPAddr, code, message string) {
	packet := &Packet{
		Type:  PacketTypeError,
		Error: message,
		Payload: map[string]string{
			"code":    code,
			"message": message,
		},
	}
	s.sendPacket(addr, packet)
}

// cleanupStaleClients removes clients that haven't been seen recently
func (s *Server) cleanupStaleClients(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			staleThreshold := 30 * time.Minute

			var toRemove []string
			for key, client := range s.clients {
				client.mu.RLock()
				lastSeen := client.LastSeen
				client.mu.RUnlock()

				if now.Sub(lastSeen) > staleThreshold {
					toRemove = append(toRemove, key)
				}
			}
			s.mu.Unlock()

			// Remove stale clients
			for _, key := range toRemove {
				log.Printf("Removing stale client: %s", key)
				s.removeClient(key)
			}
		}
	}
}

// recordSubscription records the subscription in the database
func (s *Server) recordSubscription(ctx context.Context, client *Client) {
	if s.db == nil {
		return
	}

	client.mu.RLock()
	userID := client.UserID
	novelIDs := client.NovelIDs
	allNovels := client.AllNovels
	client.mu.RUnlock()

	if allNovels {
		// Subscribe to all novels - insert a single record with NULL novel_id
		_, err := s.db.ExecContext(ctx, `
			INSERT OR REPLACE INTO Notification_Subscriptions (User_Id, Novel_Id, Is_Active, Created_At)
			VALUES (?, NULL, 1, ?)
		`, userID, time.Now())
		if err != nil {
			log.Printf("Error recording subscription: %v", err)
		}
	} else {
		// Subscribe to specific novels
		for _, novelID := range novelIDs {
			_, err := s.db.ExecContext(ctx, `
				INSERT OR REPLACE INTO Notification_Subscriptions (User_Id, Novel_Id, Is_Active, Created_At)
				VALUES (?, ?, 1, ?)
			`, userID, novelID, time.Now())
			if err != nil {
				log.Printf("Error recording subscription: %v", err)
			}
		}
	}
}

// GetClientCount returns the current number of registered clients
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}
