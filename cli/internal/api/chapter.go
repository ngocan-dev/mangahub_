package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

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

// GetChapter retrieves a chapter by its identifier.
func (c *Client) GetChapter(ctx context.Context, chapterID int64) (*Chapter, error) {
	var resp Chapter
	endpoint := fmt.Sprintf("/chapters/%d", chapterID)
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
