package favorite

import (
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
)

// LibraryStatus describes how a manga appears in user's library
type LibraryStatus struct {
	Status      string     `json:"status"`
	IsFavorite  bool       `json:"is_favorite"`
	Rating      *int       `json:"rating,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AddToLibraryRequest holds payload for adding manga to library
type AddToLibraryRequest struct {
	Status         string `json:"status" binding:"required"`
	CurrentChapter int    `json:"current_chapter"`
	IsFavorite     bool   `json:"is_favorite"`
}

// AddToLibraryResponse returns result of library addition
type AddToLibraryResponse struct {
	Message       string                `json:"message"`
	LibraryStatus *LibraryStatus        `json:"library_status"`
	UserProgress  *history.UserProgress `json:"user_progress,omitempty"`
}
