package friend

import "time"

// UserSummary contains public user info for friend features
type UserSummary struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// Friendship represents a friend relationship or request
type Friendship struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	FriendID   int64      `json:"friend_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}
