package websocket

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ngocan-dev/mangahub/backend/domain/friend"
)

// DirectChatHub manages:
// - presence connections: presence[userID][connID] => *websocket.Conn
// - direct chat connections: chats[userID][friendID] => *websocket.Conn
type DirectChatHub struct {
	db         *sql.DB
	friendRepo *friend.Repository

	mu       sync.RWMutex
	presence map[int64]map[int64]*websocket.Conn
	chats    map[int64]map[int64]*websocket.Conn

	connSeq int64
}

// DirectMessage represents a chat payload exchanged between friends.
type DirectMessage struct {
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Nếu bạn cần strict origin: check ở đây.
		return true
	},
}

// NewDirectChatHub constructs a new hub.
func NewDirectChatHub(db *sql.DB) *DirectChatHub {
	return &DirectChatHub{
		db:         db,
		friendRepo: friend.NewRepository(db),
		presence:   make(map[int64]map[int64]*websocket.Conn),
		chats:      make(map[int64]map[int64]*websocket.Conn),
	}
}

// -----------------------
// Presence (online users)
// -----------------------

func (h *DirectChatHub) GetOnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]int64, 0, len(h.presence))
	for userID := range h.presence {
		ids = append(ids, userID)
	}
	return ids
}

func (h *DirectChatHub) addPresenceConn(userID int64, conn *websocket.Conn) int64 {
	connID := atomic.AddInt64(&h.connSeq, 1)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.presence[userID] == nil {
		h.presence[userID] = make(map[int64]*websocket.Conn)
	}
	h.presence[userID][connID] = conn
	return connID
}

func (h *DirectChatHub) removePresenceConn(userID, connID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.presence[userID]; ok {
		if c, exists := conns[connID]; exists {
			_ = c.Close()
		}
		delete(conns, connID)
		if len(conns) == 0 {
			delete(h.presence, userID)
		}
	}
}

func (h *DirectChatHub) broadcastPresence() {
	online := h.GetOnlineUserIDs()
	onlineSet := make(map[int64]struct{}, len(online))
	for _, id := range online {
		onlineSet[id] = struct{}{}
	}

	// Snapshot current connections to avoid holding locks while hitting DB
	h.mu.RLock()
	connsByUser := make(map[int64][]*websocket.Conn, len(h.presence))
	for uid, userConns := range h.presence {
		list := make([]*websocket.Conn, 0, len(userConns))
		for _, conn := range userConns {
			list = append(list, conn)
		}
		connsByUser[uid] = list
	}
	h.mu.RUnlock()

	for uid, conns := range connsByUser {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		friends, err := h.friendRepo.ListFriends(ctx, uid)
		cancel()
		if err != nil {
			log.Printf("[presence] list friends failed user_id=%d err=%v", uid, err)
			continue
		}

		visible := make([]int64, 0, len(friends))
		for _, f := range friends {
			if _, ok := onlineSet[f.ID]; ok {
				visible = append(visible, f.ID)
			}
		}

		msg := map[string]any{
			"type":            "presence:update",
			"online_user_ids": visible,
		}
		for _, conn := range conns {
			_ = conn.WriteJSON(msg)
		}
	}

	log.Printf("[presence] online=%v", online)
}

// -----------------------
// Direct chat connections
// -----------------------

func (h *DirectChatHub) setChatConn(userID, friendID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chats[userID] == nil {
		h.chats[userID] = make(map[int64]*websocket.Conn)
	}
	h.chats[userID][friendID] = conn
}

func (h *DirectChatHub) removeChatConn(userID, friendID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if friends, ok := h.chats[userID]; ok {
		if c, exists := friends[friendID]; exists {
			_ = c.Close()
		}
		delete(friends, friendID)
		if len(friends) == 0 {
			delete(h.chats, userID)
		}
	}
}

func (h *DirectChatHub) getChatConn(userID, friendID int64) *websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if friends, ok := h.chats[userID]; ok {
		return friends[friendID]
	}
	return nil
}

// -----------------------
// Public entry point
// -----------------------

// HandleWS upgrades websocket.
// - If friend_id is empty -> presence-only connection: receives presence:update broadcasts.
// - If friend_id is present -> direct chat, requires friendship check.
func (h *DirectChatHub) HandleWS(w http.ResponseWriter, r *http.Request, userID int64) {
	friendIDStr := r.URL.Query().Get("friend_id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	// Always register presence for this WS connection
	pConnID := h.addPresenceConn(userID, conn)
	h.broadcastPresence()

	// Ensure cleanup
	defer func() {
		h.removePresenceConn(userID, pConnID)
		h.broadcastPresence()
	}()

	// Presence-only mode
	if friendIDStr == "" {
		log.Printf("[presence] connected user_id=%d conn_id=%d", userID, pConnID)
		h.presenceReadLoop(userID, pConnID, conn)
		return
	}

	// Chat mode
	friendID, err := strconv.ParseInt(friendIDStr, 10, 64)
	if err != nil || friendID <= 0 {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid friend_id"))
		return
	}
	if friendID == userID {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "cannot chat with yourself"))
		return
	}

	if !h.areFriends(r.Context(), userID, friendID) {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "not friends"))
		return
	}

	// Track chat connection (per friendID)
	h.setChatConn(userID, friendID, conn)
	log.Printf("[chat] connected user_id=%d friend_id=%d", userID, friendID)

	// When chat ends, remove chat conn (presence cleanup handled by defer above)
	defer func() {
		h.removeChatConn(userID, friendID)
		log.Printf("[chat] disconnected user_id=%d friend_id=%d", userID, friendID)
	}()

	h.chatReadLoop(userID, friendID, conn)
}

// -----------------------
// Read loops
// -----------------------

func (h *DirectChatHub) presenceReadLoop(userID, connID int64, conn *websocket.Conn) {
	// keepalive / detect disconnect
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// ReadMessage để bắt close frame / ping/pong
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("[presence] read end user_id=%d conn_id=%d err=%v", userID, connID, err)
			return
		}
	}
}

func (h *DirectChatHub) chatReadLoop(userID, friendID int64, conn *websocket.Conn) {
	for {
		var payload struct {
			Message string `json:"message"`
		}
		if err := conn.ReadJSON(&payload); err != nil {
			log.Printf("[chat] read error user_id=%d friend_id=%d err=%v", userID, friendID, err)
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

// -----------------------
// Chat dispatch
// -----------------------

func (h *DirectChatHub) dispatchMessage(msg DirectMessage) {
	// Send to friend if connected
	if friendConn := h.getChatConn(msg.To, msg.From); friendConn != nil {
		_ = friendConn.WriteJSON(msg)
	}
	// Echo back to sender
	if senderConn := h.getChatConn(msg.From, msg.To); senderConn != nil {
		_ = senderConn.WriteJSON(msg)
	}
}

// -----------------------
// Friendship check
// -----------------------

func (h *DirectChatHub) areFriends(ctx context.Context, userID, friendID int64) bool {
	ok, err := h.friendRepo.AreFriends(ctx, userID, friendID)
	if err != nil {
		log.Printf("[chat] friend check failed user_id=%d friend_id=%d err=%v", userID, friendID, err)
		return false
	}
	return ok
}
