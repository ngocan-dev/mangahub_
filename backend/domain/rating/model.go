package rating

import "time"

// UserRating represents a user's rating for a manga.
type UserRating struct {
	UserID    int64     `json:"user_id"`
	MangaID   int64     `json:"manga_id"`
	Score     int       `json:"score"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
