package library

import (
	"time"

	"github.com/ngocan-dev/mangahub/backend/domain/history"
)

// LibraryStatus describes how a manga appears in user's library without rating/favorite metadata
type LibraryStatus struct {
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// LibraryEntry represents an item in the user's library list
type LibraryEntry struct {
	MangaID     int64      `json:"manga_id"`
	Title       string     `json:"title"`
	CoverImage  string     `json:"cover_image"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	LastUpdated time.Time  `json:"last_updated_at"`
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

// UpdateLibraryStatusRequest holds payload for updating status
type UpdateLibraryStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// GetLibraryResponse represents a full library listing
type GetLibraryResponse struct {
	Entries []LibraryEntry `json:"entries"`
}
