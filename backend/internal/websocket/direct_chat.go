package websocket

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// DirectChatHub manages 1-to-1 chat connections: map[userID]map[friendID]*Conn
type DirectChatHub struct {
	db          *sql.DB
	connections map[int64]map[int64]*websocket.Conn
	mu          sync.RWMutex
}

// DirectMessage represents a chat payload exchanged between friends.
type DirectMessage struct {
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// NewDirectChatHub constructs a new direct chat hub.
func NewDirectChatHub(db *sql.DB) *DirectChatHub {
	return &DirectChatHub{
		db:          db,
		connections: make(map[int64]map[int64]*websocket.Conn),
	}
}

// HandleDirectChat upgrades the connection and enforces friendship checks.
func (h *DirectChatHub) HandleDirectChat(w http.ResponseWriter, r *http.Request, userID int64) {
	friendIDStr := r.URL.Query().Get("friend_id")
	friendID, err := strconv.ParseInt(friendIDStr, 10, 64)
	if err != nil || friendID <= 0 {
		http.Error(w, "invalid friend_id", http.StatusBadRequest)
		return
	}
	if friendID == userID {
		http.Error(w, "cannot chat with yourself", http.StatusBadRequest)
		return
	}

	if !h.areFriends(r.Context(), userID, friendID) {
		http.Error(w, "not friends", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("direct chat upgrade error: %v", err)
		return
	}
	h.addConnection(userID, friendID, conn)
	log.Printf("direct chat connected user_id=%d friend_id=%d", userID, friendID)

	go h.readLoop(userID, friendID, conn)
}

func (h *DirectChatHub) addConnection(userID, friendID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.connections[userID]; !ok {
		h.connections[userID] = make(map[int64]*websocket.Conn)
	}
	h.connections[userID][friendID] = conn
}

func (h *DirectChatHub) removeConnection(userID, friendID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if friends, ok := h.connections[userID]; ok {
		if conn, exists := friends[friendID]; exists {
			_ = conn.Close()
		}
		delete(friends, friendID)
		if len(friends) == 0 {
			delete(h.connections, userID)
		}
	}
}

func (h *DirectChatHub) areFriends(ctx context.Context, userID, friendID int64) bool {
	var count int
	if err := h.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM friends WHERE user_id = ? AND friend_id = ?
    `, userID, friendID).Scan(&count); err != nil {
		log.Printf("direct chat friend check failed user_id=%d friend_id=%d err=%v", userID, friendID, err)
		return false
	}
	if count == 0 {
		return false
	}
	return true
}

func (h *DirectChatHub) readLoop(userID, friendID int64, conn *websocket.Conn) {
	defer h.removeConnection(userID, friendID)
	for {
		var payload struct {
			Message string `json:"message"`
		}
		if err := conn.ReadJSON(&payload); err != nil {
			log.Printf("direct chat read error user_id=%d friend_id=%d err=%v", userID, friendID, err)
			return
		}
		msg := DirectMessage{
			From:      userID,
			To:        friendID,
			Message:   payload.Message,
			Timestamp: time.Now().Unix(),
		}
		h.dispatchMessage(msg)
	}
}

func (h *DirectChatHub) dispatchMessage(msg DirectMessage) {
	// Send to friend if connected
	if friendConn := h.getConnection(msg.To, msg.From); friendConn != nil {
		_ = friendConn.WriteJSON(msg)
	}
	// Echo back to sender
	if senderConn := h.getConnection(msg.From, msg.To); senderConn != nil {
		_ = senderConn.WriteJSON(msg)
	}
}

func (h *DirectChatHub) getConnection(userID, friendID int64) *websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if friends, ok := h.connections[userID]; ok {
		if conn, exists := friends[friendID]; exists {
			return conn
		}
	}
	return nil
}
