package rating

import "time"

// UserRating represents a single rating left by a user for a manga.
type UserRating struct {
	ID      int64     `json:"id"`
	UserID  int64     `json:"user_id"`
	MangaID int64     `json:"manga_id"`
	Rating  int       `json:"rating"`
	RatedAt time.Time `json:"rated_at"`
}

// AggregateRating captures aggregate statistics for a manga's ratings.
type AggregateRating struct {
	MangaID int64   `json:"manga_id"`
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
