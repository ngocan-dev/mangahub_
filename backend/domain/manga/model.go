package manga

import (
	"github.com/ngocan-dev/mangahub/backend/domain/history"
	"github.com/ngocan-dev/mangahub/backend/domain/library"
	pkgchapter "github.com/ngocan-dev/mangahub/backend/pkg/models"
)

// Manga represents a manga/novel entity
type Manga struct {
	ID          int64   `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Artist      string  `json:"artist,omitempty"`
	Genre       string  `json:"genre"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	RatingPoint float64 `json:"rating_point"`
	Views       int64   `json:"views,omitempty"`
	// RelevanceScore represents the ranking score returned by full-text search.
	// It is omitted when not performing text search.
	RelevanceScore float64 `json:"relevance_score,omitempty"`
}

// SearchRequest represents search criteria
type SearchRequest struct {
	Query     string   `form:"q" json:"query"`
	Genres    []string `form:"genres" json:"genres"`
	Status    string   `form:"status" json:"status"`
	MinRating *float64 `form:"min_rating" json:"min_rating"`
	MaxRating *float64 `form:"max_rating" json:"max_rating"`
	YearFrom  *int     `form:"year_from" json:"year_from"`
	YearTo    *int     `form:"year_to" json:"year_to"`
	Page      int      `form:"page" json:"page"`
	Limit     int      `form:"limit" json:"limit"`
	SortBy    string   `form:"sort_by" json:"sort_by"`
}

// SearchResponse represents paginated search results
type SearchResponse struct {
	Results []Manga `json:"results"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Pages   int     `json:"pages"`
}

// MangaDetail represents detailed manga information
type MangaDetail struct {
	Manga
	ChapterCount  int                         `json:"chapter_count"`
	Chapters      []pkgchapter.ChapterSummary `json:"chapters,omitempty"`
	LibraryStatus *library.LibraryStatus      `json:"library_status,omitempty"`
	UserProgress  *history.UserProgress       `json:"user_progress,omitempty"`
}

// CreateMangaRequest captures data required to create a manga record.
type CreateMangaRequest struct {
	Title       string
	AltTitle    string
	Slug        string
	CoverURL    string
	Author      string
	Artist      string
	Status      string
	Synopsis    string
	Genres      []string
	Rating      float64
	Views       int64
	Language    string
	LastChapter int
}
