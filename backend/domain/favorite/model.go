package favorite

import "time"

// FavoriteEntry represents a user's favorite manga entry
type FavoriteEntry struct {
	MangaID    int64      `json:"manga_id"`
	Title      string     `json:"title"`
	CoverImage string     `json:"cover_image"`
	AddedAt    *time.Time `json:"added_at,omitempty"`
}

// FavoritesResponse represents a collection of favorite entries
type FavoritesResponse struct {
	Favorites []FavoriteEntry `json:"favorites"`
}

// FavoriteStatusResponse indicates whether a manga is in favorites
type FavoriteStatusResponse struct {
	IsFavorite bool `json:"is_favorite"`
}
