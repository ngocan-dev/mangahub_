package chat

import "time"

// ChatRoom represents a private room between two friends.
type ChatRoom struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	CreatedBy int64  `json:"created_by"`
}

// ChatMessage represents a persisted chat message.
type ChatMessage struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ConversationSummary is used for left-panel listings.
type ConversationSummary struct {
	FriendID       int64        `json:"friend_id"`
	FriendUsername string       `json:"friend_username"`
	FriendAvatar   *string      `json:"friend_avatar"`
	RoomID         *int64       `json:"room_id"`
	LastMessage    *ChatMessage `json:"last_message,omitempty"`
	LastMessageAt  *time.Time   `json:"last_message_at"`
}
