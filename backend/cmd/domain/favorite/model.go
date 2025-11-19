package favorite

import (
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
)

// LibraryStatus describes how a manga appears in user's library
type LibraryStatus struct {
	Status      string     `json:"status"`
	Rating      *int       `json:"rating,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AddToLibraryRequest holds payload for adding manga to library
type AddToLibraryRequest struct {
	Status         string `json:"status" binding:"required"`
	CurrentChapter int    `json:"current_chapter"`
}

// AddToLibraryResponse returns result of library addition
type AddToLibraryResponse struct {
	Message       string                `json:"message"`
	LibraryStatus *LibraryStatus        `json:"library_status"`
	UserProgress  *history.UserProgress `json:"user_progress,omitempty"`
}

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
