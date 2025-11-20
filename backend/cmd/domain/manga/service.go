package manga

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
)

var (
	ErrDatabaseError       = errors.New("database error")
	ErrMangaNotFound       = errors.New("manga not found")
	ErrDatabaseUnavailable = errors.New("database unavailable")
)

type ChapterCounter interface {
	GetChapterCount(ctx context.Context, mangaID int64) (int, error)
}

type MangaCacher interface {
	GetMangaDetail(ctx context.Context, mangaID int64) (*MangaDetail, error)
	SetMangaDetail(ctx context.Context, mangaID int64, detail *MangaDetail) error
	InvalidateMangaDetail(ctx context.Context, mangaID int64) error
	GetSearchResults(ctx context.Context, cacheKey string) (*SearchResponse, error)
	SetSearchResults(ctx context.Context, cacheKey string, response *SearchResponse) error
}

type DBHealthChecker interface {
	IsHealthy() bool
}

type Service struct {
	repo           *Repository
	cache          MangaCacher
	dbHealth       DBHealthChecker
	chapterCounter ChapterCounter
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) SetCache(c MangaCacher) {
	s.cache = c
}

func (s *Service) SetDBHealth(h DBHealthChecker) {
	s.dbHealth = h
}

func (s *Service) SetChapterService(counter ChapterCounter) {
	s.chapterCounter = counter
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Page > 10000 {
		req.Page = 10000
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var cacheKey string
	if s.cache != nil {
		cacheKey = GenerateSearchCacheKey(req)
		if cached, err := s.cache.GetSearchResults(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	results, total, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	pages := int(math.Ceil(float64(total) / float64(req.Limit)))
	if pages == 0 {
		pages = 1
	}

	resp := &SearchResponse{
		Results: results,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Pages:   pages,
	}

	if s.cache != nil {
		_ = s.cache.SetSearchResults(ctx, cacheKey, resp)
	}

	return resp, nil
}

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

func (s *Service) GetByID(ctx context.Context, mangaID int64) (*Manga, error) {
	m, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return m, nil
}

func (s *Service) GetDetails(ctx context.Context, mangaID int64, userID *int64) (*MangaDetail, error) {
	dbHealthy := s.dbHealth == nil || s.dbHealth.IsHealthy()

	if s.cache != nil && userID == nil {
		if cached, err := s.cache.GetMangaDetail(ctx, mangaID); err == nil && cached != nil {
			return cached, nil
		}
	}

	if !dbHealthy {
		if s.cache != nil {
			if cached, err := s.cache.GetMangaDetail(ctx, mangaID); err == nil && cached != nil {
				return cached, nil
			}
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
	if s.chapterCounter != nil {
		if count, err := s.chapterCounter.GetChapterCount(ctx, mangaID); err == nil {
			chapterCount = count
		}
	}

	detail := &MangaDetail{
		Manga:        *manga,
		ChapterCount: chapterCount,
	}

	if s.cache != nil {
		_ = s.cache.SetMangaDetail(ctx, mangaID, detail)
	}

	return detail, nil
}
