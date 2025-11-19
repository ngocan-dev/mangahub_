package chapter

import "time"

// ChapterSummary represents a chapter without heavy content payloads.
type ChapterSummary struct {
	ID        int64      `json:"id"`
	MangaID   int64      `json:"manga_id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Chapter represents a full chapter including its content.
type Chapter struct {
	ChapterSummary
	Content string `json:"content"`
}
