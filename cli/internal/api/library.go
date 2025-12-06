package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// LibraryEntry represents a manga entry in the user's library.
type LibraryEntry struct {
	MangaID        string  `json:"manga_id"`
	Title          string  `json:"title"`
	Status         string  `json:"status"`
	CurrentChapter int     `json:"current_chapter"`
	TotalChapters  *int    `json:"total_chapters"`
	Rating         *int    `json:"rating"`
	StartedAt      string  `json:"started_at"`
	UpdatedAt      string  `json:"updated_at"`
	CompletedAt    string  `json:"completed_at"`
	Description    *string `json:"description"`
}

// LibraryAddResponse is returned when adding a manga to the library.
type LibraryAddResponse struct {
	MangaID string `json:"manga_id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Rating  *int   `json:"rating"`
}

// LibraryUpdateResponse is returned when updating a library entry.
type LibraryUpdateResponse struct {
	MangaID string `json:"manga_id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Rating  *int   `json:"rating"`
}

// AddToLibrary adds a manga to the user's library.
func (c *Client) AddToLibrary(ctx context.Context, mangaID, status string, rating *int) (*LibraryAddResponse, error) {
	payload := map[string]any{
		"manga_id": mangaID,
		"status":   status,
	}
	if rating != nil {
		payload["rating"] = *rating
	}

	var resp LibraryAddResponse
	if err := c.doRequest(ctx, http.MethodPost, "/library/add", payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLibrary retrieves library entries with optional filters.
func (c *Client) ListLibrary(ctx context.Context, status, sortBy, order string) ([]LibraryEntry, error) {
	u, _ := url.Parse(c.baseURL + "/library/list")
	q := u.Query()
	q.Set("status", status)
	q.Set("sort_by", sortBy)
	q.Set("order", order)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := checkStatus(res); err != nil {
		return nil, err
	}

	var entries []LibraryEntry
	if err := json.NewDecoder(res.Body).Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// RemoveFromLibrary deletes a manga entry from the user's library.
func (c *Client) RemoveFromLibrary(ctx context.Context, mangaID string) error {
	endpoint := fmt.Sprintf("/library/remove/%s", url.PathEscape(mangaID))
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil)
}

// UpdateLibraryEntry updates status or rating for a manga entry.
func (c *Client) UpdateLibraryEntry(ctx context.Context, mangaID, status string, rating *int) (*LibraryUpdateResponse, error) {
	payload := map[string]any{
		"manga_id": mangaID,
		"status":   status,
	}
	if rating != nil {
		payload["rating"] = *rating
	}

	var resp LibraryUpdateResponse
	if err := c.doRequest(ctx, http.MethodPost, "/library/update", payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
