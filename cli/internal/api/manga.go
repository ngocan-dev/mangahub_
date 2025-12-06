package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// MangaSearchFilters captures optional search parameters.
type MangaSearchFilters struct {
	Genre       string
	Status      string
	Author      string
	YearFrom    int
	YearTo      int
	MinChapters int
	SortBy      string
	Order       string
	Limit       int
}

// MangaSearchResult represents a manga search item.
type MangaSearchResult struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AltTitles []string `json:"alt_titles"`
	Author    string   `json:"author"`
	Status    string   `json:"status"`
	Chapters  int      `json:"chapters"`
}

// MangaInfoResponse contains manga metadata and optional library info.
type MangaInfoResponse struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	AltTitles     []string          `json:"alt_titles"`
	Author        string            `json:"author"`
	Artist        string            `json:"artist"`
	Genres        []string          `json:"genres"`
	Status        string            `json:"status"`
	Year          int               `json:"year"`
	Chapters      int               `json:"chapters"`
	Volumes       int               `json:"volumes"`
	Serialization string            `json:"serialization"`
	Publisher     string            `json:"publisher"`
	Description   string            `json:"description"`
	Links         map[string]string `json:"links"`
	Library       *MangaLibraryInfo `json:"library,omitempty"`
}

// MangaLibraryInfo holds user-specific library data.
type MangaLibraryInfo struct {
	Status         string `json:"status"`
	CurrentChapter int    `json:"current_chapter"`
	LastUpdated    string `json:"last_updated"`
	StartedReading string `json:"started_reading"`
	Rating         int    `json:"rating"`
}

// MangaListFilters captures list filters.
type MangaListFilters struct {
	Page   int
	Limit  int
	Genre  string
	Status string
}

// MangaListItem represents a list row.
type MangaListItem struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AltTitles []string `json:"alt_titles"`
	Author    string   `json:"author"`
	Status    string   `json:"status"`
	Chapters  int      `json:"chapters"`
}

// SearchManga searches manga titles with filters.
func (c *Client) SearchManga(ctx context.Context, query string, filters MangaSearchFilters) ([]MangaSearchResult, error) {
	u, _ := url.Parse(c.baseURL + "/manga/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("genre", filters.Genre)
	q.Set("status", filters.Status)
	q.Set("author", filters.Author)
	if filters.YearFrom > 0 {
		q.Set("year_from", strconv.Itoa(filters.YearFrom))
	}
	if filters.YearTo > 0 {
		q.Set("year_to", strconv.Itoa(filters.YearTo))
	}
	if filters.MinChapters > 0 {
		q.Set("min_chapters", strconv.Itoa(filters.MinChapters))
	}
	q.Set("sort_by", filters.SortBy)
	q.Set("order", filters.Order)
	if filters.Limit > 0 {
		q.Set("limit", strconv.Itoa(filters.Limit))
	}
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

	var results []MangaSearchResult
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetMangaInfo retrieves detailed manga information.
func (c *Client) GetMangaInfo(ctx context.Context, id string) (*MangaInfoResponse, error) {
	endpoint := fmt.Sprintf("/manga/%s", url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
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

	var info MangaInfoResponse
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ListManga retrieves paginated manga list.
func (c *Client) ListManga(ctx context.Context, filters MangaListFilters) ([]MangaListItem, error) {
	u, _ := url.Parse(c.baseURL + "/manga/list")
	q := u.Query()
	if filters.Page > 0 {
		q.Set("page", strconv.Itoa(filters.Page))
	}
	if filters.Limit > 0 {
		q.Set("limit", strconv.Itoa(filters.Limit))
	}
	q.Set("genre", filters.Genre)
	q.Set("status", strings.ToLower(filters.Status))
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

	var items []MangaListItem
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}
