package tcp

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidMessage = errors.New("invalid message format")
	ErrUnauthorized   = errors.New("authentication failed")
)

// MessageType represents the type of TCP message
type MessageType string

const (
	MessageTypeAuth      MessageType = "auth"
	MessageTypeAuthResp  MessageType = "auth_response"
	MessageTypeProgress  MessageType = "progress"
	MessageTypeError     MessageType = "error"
	MessageTypeHeartbeat MessageType = "heartbeat"
)

// Message represents a TCP protocol message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AuthRequest represents authentication request from client
type AuthRequest struct {
	Token      string `json:"token"` // JWT token
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
}

// AuthResponse represents authentication response from server
type AuthResponse struct {
	Success  bool   `json:"success"`
	UserID   int64  `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ErrorResponse represents an error message
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ProgressUpdate represents a progress update broadcast
type ProgressUpdate struct {
	UserID    int64  `json:"user_id"`
	NovelID   int64  `json:"novel_id"`
	Chapter   int    `json:"chapter"`
	ChapterID *int64 `json:"chapter_id,omitempty"`
	Timestamp string `json:"timestamp"`
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
