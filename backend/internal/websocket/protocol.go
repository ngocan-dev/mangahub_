package websocket

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidMessage = errors.New("invalid message format")
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeJoin        MessageType = "join"
	MessageTypeJoined      MessageType = "joined"
	MessageTypeReconnect   MessageType = "reconnect"
	MessageTypeReconnected MessageType = "reconnected"
	MessageTypeLeave       MessageType = "leave"
	MessageTypeLeft        MessageType = "left"
	MessageTypeMessage     MessageType = "message"
	MessageTypeHistory     MessageType = "history"
	MessageTypeError       MessageType = "error"
	MessageTypeUserList    MessageType = "user_list"
	MessageTypeHeartbeat   MessageType = "heartbeat"
)

// Message represents a WebSocket message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JoinRequest represents a join request
type JoinRequest struct {
	Token    string `json:"token"`               // JWT token for authentication
	RoomID   int64  `json:"room_id"`             // Optional: specific room ID
	RoomCode string `json:"room_code,omitempty"` // Optional: room code
}

// JoinResponse represents a successful join response
type JoinResponse struct {
	Success  bool   `json:"success"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	RoomID   int64  `json:"room_id"`
	RoomName string `json:"room_name"`
	Message  string `json:"message"`
}

// ReconnectRequest represents a reconnect attempt where the client wants to resume a session
type ReconnectRequest struct {
	Token         string `json:"token"`
	RoomID        int64  `json:"room_id"`
	RoomCode      string `json:"room_code,omitempty"`
	LastMessageID int64  `json:"last_message_id"`
	Limit         int    `json:"limit,omitempty"`
}

// ReconnectResponse confirms a successful reconnection and provides context
type ReconnectResponse struct {
	Success     bool   `json:"success"`
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	RoomID      int64  `json:"room_id"`
	RoomName    string `json:"room_name"`
	Message     string `json:"message"`
	Reconnected bool   `json:"reconnected"`
}

// UserJoinedNotification represents a user join notification
type UserJoinedNotification struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	RoomID      int64  `json:"room_id"`
	Timestamp   string `json:"timestamp"`
	Reconnected bool   `json:"reconnected,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	MessageID int64  `json:"message_id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	RoomID    int64  `json:"room_id"`
	Timestamp string `json:"timestamp"`
}

// HistoryResponse represents chat history
type HistoryResponse struct {
	RoomID   int64         `json:"room_id"`
	Messages []ChatMessage `json:"messages"`
	Limit    int           `json:"limit"`
}

// UserListResponse represents list of active users
type UserListResponse struct {
	RoomID int64      `json:"room_id"`
	Users  []UserInfo `json:"users"`
}

// UserInfo represents user information
type UserInfo struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// ParseMessage parses a JSON message from bytes
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, ErrInvalidMessage
	}
	return &msg, nil
}

// SerializeMessage serializes a message to JSON bytes
func SerializeMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}

// FormatTimestamp formats a time to RFC3339 string
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
