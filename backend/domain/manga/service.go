package manga

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	pkgchapter "github.com/ngocan-dev/mangahub/backend/pkg/models"
)

var (
	ErrDatabaseError       = errors.New("database error")
	ErrMangaNotFound       = errors.New("manga not found")
	ErrDatabaseUnavailable = errors.New("database unavailable")
)

const defaultChapterListLimit = 100

// ChapterService exposes chapter operations required by the manga service
type ChapterService interface {
	GetChapterCount(ctx context.Context, mangaID int64) (int, error)
	GetChapters(ctx context.Context, mangaID int64, limit, offset int) ([]pkgchapter.ChapterSummary, error)
}

// Service provides manga metadata operations
type Service struct {
	repo           *Repository
	cache          MangaCacher
	dbHealth       DBHealthChecker
	writeQueue     WriteQueue
	chapterService ChapterService
}

// MangaCacher interface for manga caching
type MangaCacher interface {
	GetMangaDetail(ctx context.Context, mangaID int64) (*MangaDetail, error)
	SetMangaDetail(ctx context.Context, mangaID int64, detail *MangaDetail) error
	InvalidateMangaDetail(ctx context.Context, mangaID int64) error
	GetSearchResults(ctx context.Context, cacheKey string) (*SearchResponse, error)
	SetSearchResults(ctx context.Context, cacheKey string, response *SearchResponse) error
	GetPopularManga(ctx context.Context, limit int) ([]Manga, error)
	SetPopularManga(ctx context.Context, limit int, popular []Manga) error
}

// DBHealthChecker exposes database status
type DBHealthChecker interface {
	IsHealthy() bool
}

// WriteQueue defines the queuing operations used by the manga service
type WriteQueue interface {
	Enqueue(opType string, userID, mangaID int64, data map[string]interface{}) error
}

// NewService creates a manga service
func NewService(db *sql.DB) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

// SetCache configures cache backend
func (s *Service) SetCache(cache MangaCacher) {
	s.cache = cache
}

// SetDBHealth sets health checker
func (s *Service) SetDBHealth(checker DBHealthChecker) {
	s.dbHealth = checker
}

// SetWriteQueue sets the write queue used for offline write queuing
func (s *Service) SetWriteQueue(queue WriteQueue) {
	s.writeQueue = queue
}

// IsDBHealthy reports whether the database is currently healthy
func (s *Service) IsDBHealthy() bool {
	return s.dbHealth == nil || s.dbHealth.IsHealthy()
}

// SetChapterService injects the chapter service
func (s *Service) SetChapterService(chapterSvc ChapterService) {
	s.chapterService = chapterSvc
}

// Search searches for manga based on criteria
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page > 10000 {
		req.Page = 10000
	}

	dbHealthy := s.IsDBHealthy()

	if s.cache != nil {
		cacheKey := GenerateSearchCacheKey(req)
		if cached, err := s.cache.GetSearchResults(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	if !dbHealthy {
		return nil, ErrDatabaseUnavailable
	}

	results, total, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	pages := int(math.Ceil(float64(total) / float64(req.Limit)))
	if pages == 0 {
		pages = 1
	}

	response := &SearchResponse{
		Results: results,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Pages:   pages,
	}

	if s.cache != nil {
		cacheKey := GenerateSearchCacheKey(req)
		_ = s.cache.SetSearchResults(ctx, cacheKey, response)
	}

	return response, nil
}

// GenerateSearchCacheKey generates a cache key for search request
func GenerateSearchCacheKey(req SearchRequest) string {
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

// GetPopularManga returns the top-rated manga list with caching support
// Step 1: System identifies frequently requested manga
// Step 4: Subsequent requests serve data from cache
func (s *Service) GetPopularManga(ctx context.Context, limit int) ([]Manga, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	dbHealthy := s.IsDBHealthy()

	if s.cache != nil {
		if cached, err := s.cache.GetPopularManga(ctx, limit); err == nil && cached != nil {
			return cached, nil
		}
	}

	if !dbHealthy {
		return nil, ErrDatabaseUnavailable
	}

	popular, err := s.repo.GetPopularManga(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if s.cache != nil {
		_ = s.cache.SetPopularManga(ctx, limit, popular)
	}

	return popular, nil
}

// GetByID retrieves a manga entity
func (s *Service) GetByID(ctx context.Context, mangaID int64) (*Manga, error) {
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return manga, nil
}

// Exists reports whether a manga exists by ID
func (s *Service) Exists(ctx context.Context, mangaID int64) (bool, error) {
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return manga != nil, nil
}

// GetDetails retrieves detailed manga information
func (s *Service) GetDetails(ctx context.Context, mangaID int64, userID *int64) (*MangaDetail, error) {
	dbHealthy := s.IsDBHealthy()

	if s.cache != nil && userID == nil {
		if cached, err := s.cache.GetMangaDetail(ctx, mangaID); err == nil && cached != nil {
			return cached, nil
		}
	}

	if !dbHealthy && s.cache != nil {
		if cached, err := s.cache.GetMangaDetail(ctx, mangaID); err == nil && cached != nil {
			return cached, nil
		}
		return nil, ErrDatabaseUnavailable
	}

	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		if s.cache != nil {
			if cached, cacheErr := s.cache.GetMangaDetail(ctx, mangaID); cacheErr == nil && cached != nil {
				return cached, nil
			}
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	chapterCount := 0
	var chapters []pkgchapter.ChapterSummary
	if s.chapterService != nil {
		if count, err := s.chapterService.GetChapterCount(ctx, mangaID); err == nil {
			chapterCount = count
		}
		if list, err := s.chapterService.GetChapters(ctx, mangaID, defaultChapterListLimit, 0); err == nil {
			chapters = list
		}
	}

	detail := &MangaDetail{
		Manga:        *manga,
		ChapterCount: chapterCount,
		Chapters:     chapters,
	}

	if s.cache != nil && userID == nil {
		_ = s.cache.SetMangaDetail(ctx, mangaID, detail)
	}

	return detail, nil
}
