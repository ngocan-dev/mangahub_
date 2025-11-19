package udp

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidPacket = errors.New("invalid packet format")
)

// PacketType represents the type of UDP packet
type PacketType string

const (
	PacketTypeRegister     PacketType = "register"
	PacketTypeConfirm      PacketType = "confirm"
	PacketTypeUnregister   PacketType = "unregister"
	PacketTypeNotification PacketType = "notification"
	PacketTypeError        PacketType = "error"
)

// Packet represents a UDP protocol packet
type Packet struct {
	Type    PacketType  `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterRequest represents registration request from client
type RegisterRequest struct {
	UserID    int64   `json:"user_id"`              // User ID (from JWT or direct)
	Token     string  `json:"token,omitempty"`      // Optional: JWT token for validation
	NovelIDs  []int64 `json:"novel_ids,omitempty"`  // Optional: specific novels to subscribe to
	AllNovels bool    `json:"all_novels,omitempty"` // Subscribe to all novels
	DeviceID  string  `json:"device_id,omitempty"`  // Optional: device identifier
}

// RegisterResponse represents registration confirmation
type RegisterResponse struct {
	Success  bool    `json:"success"`
	UserID   int64   `json:"user_id"`
	Message  string  `json:"message,omitempty"`
	NovelIDs []int64 `json:"novel_ids,omitempty"`
	Error    string  `json:"error,omitempty"`
}

// NotificationPacket represents a chapter release notification
type NotificationPacket struct {
	NovelID   int64  `json:"novel_id"`
	NovelName string `json:"novel_name"`
	Chapter   int    `json:"chapter"`
	ChapterID int64  `json:"chapter_id"`
	Timestamp string `json:"timestamp"`
}

// ParsePacket parses a JSON packet from bytes
func ParsePacket(data []byte) (*Packet, error) {
	var packet Packet
	if err := json.Unmarshal(data, &packet); err != nil {
		return nil, ErrInvalidPacket
	}
	return &packet, nil
}

// SerializePacket serializes a packet to JSON bytes
func SerializePacket(packet *Packet) ([]byte, error) {
	return json.Marshal(packet)
}
