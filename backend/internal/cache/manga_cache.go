package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/manga"
)

const (
	// Cache key prefixes
	mangaDetailPrefix = "manga:detail:"
	mangaSearchPrefix = "manga:search:"

	// Cache expiration times
	mangaDetailExpiration = 1 * time.Hour    // Manga details cached for 1 hour
	mangaSearchExpiration = 30 * time.Minute // Search results cached for 30 minutes
)

// MangaCache provides caching for manga data
type MangaCache struct {
	client *Client
}

// NewMangaCache creates a new manga cache
func NewMangaCache(client *Client) *MangaCache {
	return &MangaCache{client: client}
}

// GetMangaDetail retrieves manga detail from cache
func (c *MangaCache) GetMangaDetail(ctx context.Context, mangaID int64) (*manga.MangaDetail, error) {
	key := fmt.Sprintf("%s%d", mangaDetailPrefix, mangaID)

	data, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil // Not in cache
	}

	var detail manga.MangaDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manga detail: %w", err)
	}

	return &detail, nil
}

// SetMangaDetail stores manga detail in cache
// Step 2: System stores manga details in Redis cache
// Step 3: System sets appropriate cache expiration times
func (c *MangaCache) SetMangaDetail(ctx context.Context, mangaID int64, detail *manga.MangaDetail) error {
	key := fmt.Sprintf("%s%d", mangaDetailPrefix, mangaID)
	return c.client.Set(ctx, key, detail, mangaDetailExpiration)
}

// InvalidateMangaDetail removes manga detail from cache
// Step 5: System updates cache when data changes
func (c *MangaCache) InvalidateMangaDetail(ctx context.Context, mangaID int64) error {
	key := fmt.Sprintf("%s%d", mangaDetailPrefix, mangaID)
	return c.client.Delete(ctx, key)
}

// InvalidateAllMangaDetails removes all manga detail caches
func (c *MangaCache) InvalidateAllMangaDetails(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", mangaDetailPrefix)
	return c.client.DeletePattern(ctx, pattern)
}

// GetSearchResults retrieves search results from cache
func (c *MangaCache) GetSearchResults(ctx context.Context, cacheKey string) (*manga.SearchResponse, error) {
	key := fmt.Sprintf("%s%s", mangaSearchPrefix, cacheKey)

	data, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil // Not in cache
	}

	var response manga.SearchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return &response, nil
}

// SetSearchResults stores search results in cache
func (c *MangaCache) SetSearchResults(ctx context.Context, cacheKey string, response *manga.SearchResponse) error {
	key := fmt.Sprintf("%s%s", mangaSearchPrefix, cacheKey)
	return c.client.Set(ctx, key, response, mangaSearchExpiration)
}

// GenerateSearchCacheKey generates a cache key for search request
func GenerateSearchCacheKey(req manga.SearchRequest) string {
	// Create a unique key based on search parameters
	key := fmt.Sprintf("q:%s", req.Query)
	if len(req.Genres) > 0 {
		key += fmt.Sprintf(":genres:%v", req.Genres)
	}
	if req.Status != "" {
		key += fmt.Sprintf(":status:%s", req.Status)
	}
	if req.MinRating != nil {
		key += fmt.Sprintf(":min_rating:%.1f", *req.MinRating)
	}
	if req.MaxRating != nil {
		key += fmt.Sprintf(":max_rating:%.1f", *req.MaxRating)
	}
	if req.YearFrom != nil {
		key += fmt.Sprintf(":year_from:%d", *req.YearFrom)
	}
	if req.YearTo != nil {
		key += fmt.Sprintf(":year_to:%d", *req.YearTo)
	}
	key += fmt.Sprintf(":page:%d:limit:%d:sort:%s", req.Page, req.Limit, req.SortBy)
	return key
}
