package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/internal/auth"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/security"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients by room
	rooms map[int64]map[*Client]bool

	// All registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Database connection
	db *sql.DB

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new hub instance
// TCP and WebSocket connections remain stable
func NewHub(db *sql.DB) *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 1000), // Increased buffer for 50-100 concurrent users
		register:   make(chan *Client, 100), // Buffered channels to prevent blocking
		unregister: make(chan *Client, 100),
		db:         db,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			// Client registered - already handled in join
			log.Printf("Client registered: UserID=%d", client.GetUserID())
		case client := <-h.unregister:
			// Main Success Scenario:
			// 1. System detects WebSocket connection closure
			// 2. Server removes connection from active list
			// 3. Server broadcasts user leave notification
			// 4. Other users see updated participant list
			// 5. Connection resources are cleaned up
			h.handleClientDisconnect(client)
		case message := <-h.broadcast:
			// Broadcast handled per message type
			// This channel can be used for general broadcasts if needed
		}
	}
}

// handleMessage handles incoming messages from clients
func (h *Hub) handleMessage(client *Client, msg *Message) {
	switch msg.Type {
	case MessageTypeJoin:
		h.handleJoin(client, msg)
	case MessageTypeReconnect:
		h.handleReconnect(client, msg)
	case MessageTypeMessage:
		h.handleChatMessage(client, msg)
	case MessageTypeLeave:
		h.handleLeave(client)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		client.SendError("unknown_type", "unknown message type")
	}
}

// handleJoin handles client join request
// Main Success Scenario:
// 1. User's browser initiates WebSocket connection
// 2. Server upgrades HTTP connection to WebSocket
// 3. Client sends join message with user credentials
// 4. Server validates user and adds to active connections
// 5. Server broadcasts user join notification to other users
// 6. User receives recent chat history
func (h *Hub) handleJoin(client *Client, msg *Message) {
	// Parse join request
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		client.SendError("invalid_request", "invalid join request")
		return
	}

	var req JoinRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		client.SendError("invalid_request", "invalid join request format")
		return
	}

	// Step 4: Validate user
	if req.Token == "" {
		client.SendError("auth_required", "authentication token required")
		return
	}

	// Validate token
	// Invalid tokens are rejected
	// Expired tokens trigger reauthentication
	claims, err := auth.ValidateToken(req.Token)
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
		return
	}
	if claims == nil {
		// Unauthorized access is prevented
		client.SendError("auth_failed", "token validation failed")
		return
	}

	// Determine room ID
	roomID, err := h.getRoomID(context.Background(), req.RoomID, req.RoomCode)
	if err != nil {
		client.SendError("room_not_found", "room not found")
		return
	}

	// Set user and room
	client.SetUser(claims.UserID, claims.Username)
	client.SetRoom(roomID)

	// Step 4: Add to active connections
	h.addClient(client, roomID)

	// Step 6: Send recent chat history
	history, err := h.getChatHistory(context.Background(), roomID, 50)
	if err == nil {
		historyMsg := &Message{
			Type:    MessageTypeHistory,
			Payload: history,
		}
		client.SendMessage(historyMsg)
	}

	// Send join confirmation
	joinResp := &Message{
		Type: MessageTypeJoined,
		Payload: JoinResponse{
			Success:  true,
			UserID:   claims.UserID,
			Username: claims.Username,
			RoomID:   roomID,
			RoomName: h.getRoomName(context.Background(), roomID),
			Message:  "joined successfully",
		},
	}
	client.SendMessage(joinResp)

	// Step 5: Broadcast user join notification to other users
	notification := &Message{
		Type: MessageTypeJoined,
		Payload: UserJoinedNotification{
			UserID:    claims.UserID,
			Username:  claims.Username,
			RoomID:    roomID,
			Timestamp: FormatTimestamp(time.Now()),
		},
	}
	h.broadcastToRoomExcept(roomID, client, notification)

	// Step 4: Send updated participant list to all users (including the new user)
	h.broadcastUserList(roomID)

	log.Printf("User joined: UserID=%d, Username=%s, RoomID=%d", claims.UserID, claims.Username, roomID)
}

// handleReconnect allows a client to resume a session and retrieve missed messages
func (h *Hub) handleReconnect(client *Client, msg *Message) {
	// Parse reconnect request
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		client.SendError("invalid_request", "invalid reconnect request")
		return
	}

	var req ReconnectRequest
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		client.SendError("invalid_request", "invalid reconnect request format")
		return
	}

	if req.Token == "" {
		client.SendError("auth_required", "authentication token required")
		return
	}

	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) {
			client.SendError("token_expired", "your session has expired. please login again")
		} else {
			client.SendError("auth_failed", "authentication failed")
		}
		return
	}

	roomID, err := h.getRoomID(context.Background(), req.RoomID, req.RoomCode)
	if err != nil {
		client.SendError("room_not_found", "room not found")
		return
	}

	client.SetUser(claims.UserID, claims.Username)
	client.SetRoom(roomID)
	h.addClient(client, roomID)

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	if history, err := h.getChatHistorySince(context.Background(), roomID, req.LastMessageID, limit); err == nil {
		historyMsg := &Message{
			Type:    MessageTypeHistory,
			Payload: history,
		}
		client.SendMessage(historyMsg)
	}

	reconnectResp := &Message{
		Type: MessageTypeReconnected,
		Payload: ReconnectResponse{
			Success:     true,
			UserID:      claims.UserID,
			Username:    claims.Username,
			RoomID:      roomID,
			RoomName:    h.getRoomName(context.Background(), roomID),
			Message:     "reconnected successfully",
			Reconnected: true,
		},
	}
	client.SendMessage(reconnectResp)

	notification := &Message{
		Type: MessageTypeJoined,
		Payload: UserJoinedNotification{
			UserID:      claims.UserID,
			Username:    claims.Username,
			RoomID:      roomID,
			Timestamp:   FormatTimestamp(time.Now()),
			Reconnected: true,
		},
	}
	h.broadcastToRoomExcept(roomID, client, notification)
	h.broadcastUserList(roomID)

	log.Printf("User reconnected: UserID=%d, Username=%s, RoomID=%d, LastMessageID=%d", claims.UserID, claims.Username, roomID, req.LastMessageID)
}

// handleChatMessage handles chat messages
// Main Success Scenario:
// 1. User types message and clicks send
// 2. Client sends message via WebSocket connection
// 3. Server receives message and validates user
// 4. Server broadcasts message to all connected clients
// 5. All users receive and display the message
func (h *Hub) handleChatMessage(client *Client, msg *Message) {
	// Step 3: Validate user
	userID := client.GetUserID()
	if userID == 0 {
		// A2: User not authenticated - Server rejects message
		client.SendError("not_authenticated", "user not authenticated")
		return
	}

	roomID := client.GetRoomID()
	if roomID == 0 {
		client.SendError("not_in_room", "user not in a room")
		return
	}

	// Parse message payload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		client.SendError("invalid_request", "invalid message format")
		return
	}

	var chatMsg struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(payloadBytes, &chatMsg); err != nil {
		client.SendError("invalid_request", "invalid message payload")
		return
	}

	// A1: Message too long - Server returns error to sender
	// Input length limits are enforced
	// XSS attempts are sanitized
	if err := security.ValidateLength(chatMsg.Content, 1, security.MaxMessageLength); err != nil {
		if errors.Is(err, security.ErrInputTooShort) {
			client.SendError("empty_message", "message cannot be empty")
			return
		}
		if errors.Is(err, security.ErrInputTooLong) {
			client.SendError("message_too_long", "message exceeds maximum length")
			return
		}
		client.SendError("invalid_message", "invalid message format")
		return
	}

	// Check for SQL injection and XSS
	// SQL injection attempts are blocked
	if err := security.DetectSQLInjection(chatMsg.Content); err != nil {
		client.SendError("invalid_message", "message contains invalid content")
		return
	}

	// Sanitize message content to prevent XSS
	// XSS attempts are sanitized
	sanitizedContent := security.SanitizeString(chatMsg.Content)

	// Save message to database (use sanitized content)
	messageID, err := h.saveMessage(context.Background(), roomID, userID, sanitizedContent)
	if err != nil {
		log.Printf("Error saving message to database: %v", err)
		client.SendError("database_error", "failed to save message")
		return
	}

	// Create chat message object
	username := client.GetUsername()
	chatMessage := ChatMessage{
		MessageID: messageID,
		UserID:    userID,
		Username:  username,
		Content:   sanitizedContent,
		RoomID:    roomID,
		Timestamp: FormatTimestamp(time.Now()),
	}

	// Step 4: Broadcast message to all connected clients
	broadcastMsg := &Message{
		Type:    MessageTypeMessage,
		Payload: chatMessage,
	}
	h.broadcastToRoom(roomID, broadcastMsg)

	log.Printf("Message sent: UserID=%d, Username=%s, RoomID=%d, MessageID=%d",
		userID, username, roomID, messageID)
}

// handleLeave handles explicit client leave request (user sends leave message)
func (h *Hub) handleLeave(client *Client) {
	userID := client.GetUserID()
	username := client.GetUsername()
	roomID := client.GetRoomID()

	if roomID > 0 && userID > 0 {
		// Broadcast leave notification before removing
		notification := &Message{
			Type: MessageTypeLeft,
			Payload: map[string]interface{}{
				"user_id":   userID,
				"username":  username,
				"room_id":   roomID,
				"timestamp": FormatTimestamp(time.Now()),
			},
		}
		h.broadcastToRoomExcept(roomID, client, notification)

		// Send updated participant list
		h.broadcastUserList(roomID)

		// Remove client after broadcasting
		h.removeClient(client)

		log.Printf("User left: UserID=%d, Username=%s, RoomID=%d", userID, username, roomID)
	}
}

// addClient adds a client to the hub
func (h *Hub) addClient(client *Client, roomID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to all clients
	h.clients[client] = true

	// Add to room
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
}

// handleClientDisconnect handles client disconnection
// Main Success Scenario:
// 1. System detects WebSocket connection closure
// 2. Server removes connection from active list
// 3. Server broadcasts user leave notification
// 4. Other users see updated participant list
// 5. Connection resources are cleaned up
func (h *Hub) handleClientDisconnect(client *Client) {
	userID := client.GetUserID()
	username := client.GetUsername()
	roomID := client.GetRoomID()

	// Only broadcast if user was authenticated and in a room
	if userID > 0 && roomID > 0 {
		// Step 3: Broadcast user leave notification
		leaveNotification := &Message{
			Type: MessageTypeLeft,
			Payload: map[string]interface{}{
				"user_id":   userID,
				"username":  username,
				"room_id":   roomID,
				"timestamp": FormatTimestamp(time.Now()),
			},
		}
		h.broadcastToRoomExcept(roomID, client, leaveNotification)

		// Step 4: Send updated participant list to remaining users
		h.broadcastUserList(roomID)
	}

	// Step 2: Remove connection from active list
	h.removeClient(client)

	// Step 5: Connection resources are cleaned up (handled by defer in ReadPump)
	log.Printf("Client disconnected: UserID=%d, Username=%s, RoomID=%d", userID, username, roomID)
}

// removeClient removes a client from the hub (internal, called after broadcasting)
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from all clients
	delete(h.clients, client)

	// Remove from room
	roomID := client.GetRoomID()
	if room, ok := h.rooms[roomID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

// broadcastUserList broadcasts updated user list to all clients in a room
func (h *Hub) broadcastUserList(roomID int64) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// Build user list
	users := make([]UserInfo, 0, len(room))
	for client := range room {
		if client.GetUserID() > 0 {
			users = append(users, UserInfo{
				UserID:   client.GetUserID(),
				Username: client.GetUsername(),
			})
		}
	}
	h.mu.RUnlock()

	// Broadcast user list
	userListMsg := &Message{
		Type: MessageTypeUserList,
		Payload: UserListResponse{
			RoomID: roomID,
			Users:  users,
		},
	}
	h.broadcastToRoom(roomID, userListMsg)
}

// broadcastToRoom broadcasts a message to all clients in a room
func (h *Hub) broadcastToRoom(roomID int64, msg *Message) {
	data, err := SerializeMessage(msg)
	if err != nil {
		log.Printf("Error serializing message: %v", err)
		return
	}

	h.mu.RLock()
	room, exists := h.rooms[roomID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := make([]*Client, 0, len(room))
	for client := range room {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			// Client send buffer full
			log.Printf("Client send buffer full, dropping message")
		}
	}
}

// broadcastToRoomExcept broadcasts a message to all clients in a room except one
func (h *Hub) broadcastToRoomExcept(roomID int64, except *Client, msg *Message) {
	data, err := SerializeMessage(msg)
	if err != nil {
		log.Printf("Error serializing message: %v", err)
		return
	}

	h.mu.RLock()
	room, exists := h.rooms[roomID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := make([]*Client, 0, len(room))
	for client := range room {
		if client != except {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			// Client send buffer full
			log.Printf("Client send buffer full, dropping message")
		}
	}
}

// getRoomID gets room ID from room ID or room code
func (h *Hub) getRoomID(ctx context.Context, roomID int64, roomCode string) (int64, error) {
	if roomID > 0 {
		// Verify room exists
		var exists int
		err := h.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM Chat_Rooms WHERE Room_Id = ?
		`, roomID).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if exists > 0 {
			return roomID, nil
		}
	}

	if roomCode != "" {
		// Get room by code
		var id int64
		err := h.db.QueryRowContext(ctx, `
			SELECT Room_Id FROM Chat_Rooms WHERE Room_Code = ?
		`, roomCode).Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}

	// Default to general room (room_id = 1) or get/create it
	var id int64
	err := h.db.QueryRowContext(ctx, `
		SELECT Room_Id FROM Chat_Rooms WHERE Room_Code = 'general' OR Room_Id = 1 LIMIT 1
	`).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Create default general room if it doesn't exist
	result, err := h.db.ExecContext(ctx, `
		INSERT INTO Chat_Rooms (Room_Code, Room_Name) VALUES ('general', 'General Chat')
	`)
	if err != nil {
		return 0, err
	}

	createdID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return createdID, nil
}

// getRoomName gets room name
func (h *Hub) getRoomName(ctx context.Context, roomID int64) string {
	var name string
	err := h.db.QueryRowContext(ctx, `
		SELECT Room_Name FROM Chat_Rooms WHERE Room_Id = ?
	`, roomID).Scan(&name)
	if err != nil {
		return "General Chat"
	}
	return name
}

// getChatHistory retrieves recent chat history
func (h *Hub) getChatHistory(ctx context.Context, roomID int64, limit int) (*HistoryResponse, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT 
			cm.Message_Id,
			cm.User_Id,
			u.Username,
			cm.Content,
			cm.Created_At
		FROM Chat_Messages cm
		JOIN Users u ON cm.User_Id = u.UserId
		WHERE cm.Room_Id = ?
		ORDER BY cm.Created_At DESC
		LIMIT ?
	`, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var msg ChatMessage
		var createdAt time.Time
		err := rows.Scan(&msg.MessageID, &msg.UserID, &msg.Username, &msg.Content, &createdAt)
		if err != nil {
			continue
		}
		msg.RoomID = roomID
		msg.Timestamp = FormatTimestamp(createdAt)
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return &HistoryResponse{
		RoomID:   roomID,
		Messages: messages,
		Limit:    limit,
	}, nil
}

// getChatHistorySince fetches chat history after a specific message ID
func (h *Hub) getChatHistorySince(ctx context.Context, roomID int64, lastMessageID int64, limit int) (*HistoryResponse, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := h.db.QueryContext(ctx, `
SELECT
cm.Message_Id,
cm.User_Id,
u.Username,
cm.Content,
cm.Created_At
FROM Chat_Messages cm
JOIN Users u ON cm.User_Id = u.UserId
WHERE cm.Room_Id = ? AND cm.Message_Id > ?
ORDER BY cm.Created_At ASC
LIMIT ?
`, roomID, lastMessageID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var msg ChatMessage
		var createdAt time.Time
		if err := rows.Scan(&msg.MessageID, &msg.UserID, &msg.Username, &msg.Content, &createdAt); err != nil {
			continue
		}
		msg.RoomID = roomID
		msg.Timestamp = FormatTimestamp(createdAt)
		messages = append(messages, msg)
	}

	return &HistoryResponse{
		RoomID:   roomID,
		Messages: messages,
		Limit:    limit,
	}, nil
}

// saveMessage saves a chat message to the database
func (h *Hub) saveMessage(ctx context.Context, roomID, userID int64, content string) (int64, error) {
	result, err := h.db.ExecContext(ctx, `
		INSERT INTO Chat_Messages (Room_Id, User_Id, Content, Created_At)
		VALUES (?, ?, ?, ?)
	`, roomID, userID, content, time.Now())
	if err != nil {
		return 0, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return messageID, nil
}
