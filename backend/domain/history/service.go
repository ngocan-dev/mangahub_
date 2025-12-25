package history

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pkgchapter "github.com/ngocan-dev/mangahub/backend/pkg/models"
)

var (
	ErrInvalidChapterNumber = errors.New("invalid chapter number")
	ErrMangaNotFound        = errors.New("manga not found")
	ErrMangaNotInLibrary    = errors.New("manga not in library")
	ErrDatabaseError        = errors.New("database error")
	ErrNoData               = errors.New("no data available")
)

// ChapterService exposes chapter operations needed by history
type ChapterService interface {
	ValidateChapter(ctx context.Context, mangaID int64, chapter int) (*pkgchapter.ChapterSummary, error)
}

// LibraryChecker verifies library membership
type LibraryChecker interface {
	CheckLibraryExists(ctx context.Context, userID, mangaID int64) (bool, error)
}

// Broadcaster broadcasts progress updates
type Broadcaster interface {
	BroadcastProgress(ctx context.Context, userID, mangaID int64, chapter int, chapterID *int64) error
}

// MangaChecker verifies manga existence
type MangaChecker interface {
	Exists(ctx context.Context, mangaID int64) (bool, error)
}

// Service manages history use cases
type Service struct {
	repo           *Repository
	chapterService ChapterService
	libraryChecker LibraryChecker
	broadcaster    Broadcaster
	mangaChecker   MangaChecker
}

// NewService builds history service
func NewService(repo *Repository, chapterSvc ChapterService, libraryChecker LibraryChecker, mangaChecker MangaChecker) *Service {
	return &Service{
		repo:           repo,
		chapterService: chapterSvc,
		libraryChecker: libraryChecker,
		mangaChecker:   mangaChecker,
	}
}

// SetBroadcaster injects optional broadcaster
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// GetProgress returns user's progress
func (s *Service) GetProgress(ctx context.Context, userID, mangaID int64) (*UserProgress, error) {
	progress, err := s.repo.GetUserProgress(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return progress, nil
}

// HasCompletedManga reports whether the user finished the manga in their library
func (s *Service) HasCompletedManga(ctx context.Context, userID, mangaID int64) (bool, error) {
	completed, err := s.repo.IsMangaCompleted(ctx, userID, mangaID)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return completed, nil
}

// UpdateProgress updates reading progress
func (s *Service) UpdateProgress(ctx context.Context, userID, mangaID int64, req UpdateProgressRequest) (*UpdateProgressResponse, error) {
	if req.CurrentChapter < 1 {
		return nil, ErrInvalidChapterNumber
	}

	if s.mangaChecker == nil {
		return nil, fmt.Errorf("%w: manga checker not configured", ErrDatabaseError)
	}

	exists, err := s.mangaChecker.Exists(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return nil, ErrMangaNotFound
	}

	inLibrary, err := s.libraryChecker.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !inLibrary {
		return nil, ErrMangaNotInLibrary
	}

	summary, err := s.chapterService.ValidateChapter(ctx, mangaID, req.CurrentChapter)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if summary == nil {
		return nil, ErrInvalidChapterNumber
	}

	var chapterID *int64
	if summary.ID != 0 {
		id := summary.ID
		chapterID = &id
	}

	if err := s.repo.UpdateProgress(ctx, userID, mangaID, req.CurrentChapter, chapterID); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	broadcasted := false
	if s.broadcaster != nil {
		if err := s.broadcaster.BroadcastProgress(ctx, userID, mangaID, req.CurrentChapter, chapterID); err == nil {
			broadcasted = true
		}
	}

	progress, err := s.repo.GetUserProgress(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	_ = s.repo.RecordActivity(ctx, userID, "READ", &mangaID, map[string]interface{}{
		"current_chapter": req.CurrentChapter,
		"chapter_id":      chapterID,
	})

	return &UpdateProgressResponse{
		Message:      "progress updated successfully",
		UserProgress: progress,
		Broadcasted:  broadcasted,
	}, nil
}

// GetFriendsActivityFeed returns friend activities
func (s *Service) GetFriendsActivityFeed(ctx context.Context, userID int64, page, limit int) (*ActivityFeedResponse, error) {
	activities, total, err := s.repo.GetFriendsActivities(ctx, userID, page, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ActivityFeedResponse{
				Activities: []Activity{},
				Total:      0,
				Page:       page,
				Limit:      limit,
				Pages:      0,
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	resp := &ActivityFeedResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}
	if len(resp.Activities) == 0 {
		resp.Activities = []Activity{}
	}
	if resp.Limit > 0 {
		resp.Pages = (resp.Total + resp.Limit - 1) / resp.Limit
	}
	return resp, nil
}

// GetReadingStatistics returns cached/calculated stats
func (s *Service) GetReadingStatistics(ctx context.Context, userID int64, force bool) (*ReadingStatistics, error) {
	if !force {
		cached, err := s.repo.GetCachedReadingStatistics(ctx, userID)
		if err == nil && cached != nil {
			if time.Since(cached.LastCalculatedAt) < time.Hour {
				return cached, nil
			}
		}
	}

	stats, err := s.repo.CalculateReadingStatistics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if stats == nil {
		return nil, ErrNoData
	}
	if stats.LastCalculatedAt.IsZero() {
		stats.LastCalculatedAt = time.Now()
	}

	if err := s.repo.SaveReadingStatistics(ctx, stats); err != nil {
		// ignore cache error
	}

	return stats, nil
}

// RecordActivity proxies to the repository to allow other services to reuse the activity feed.
func (s *Service) RecordActivity(ctx context.Context, userID int64, activityType string, mangaID *int64, payload map[string]interface{}) error {
	return s.repo.RecordActivity(ctx, userID, activityType, mangaID, payload)
}

// GetReadingAnalytics filters stats
func (s *Service) GetReadingAnalytics(ctx context.Context, userID int64, req ReadingAnalyticsRequest) (*ReadingStatistics, error) {
	if req.TimePeriod == "" {
		req.TimePeriod = "month"
	}
	stats, err := s.GetReadingStatistics(ctx, userID, false)
	if err != nil {
		return nil, err
	}

	if req.TimePeriod != "" && req.TimePeriod != "all_time" {
		switch req.TimePeriod {
		case "year":
			if len(stats.MonthlyStats) > 12 {
				stats.MonthlyStats = stats.MonthlyStats[:12]
			}
			if len(stats.YearlyStats) > 1 {
				stats.YearlyStats = stats.YearlyStats[:1]
			}
		case "month":
			if req.Year != nil && req.Month != nil {
				var filtered []MonthlyStat
				for _, m := range stats.MonthlyStats {
					if m.Year == *req.Year && m.Month == *req.Month {
						filtered = []MonthlyStat{m}
						break
					}
				}
				stats.MonthlyStats = filtered
			} else if len(stats.MonthlyStats) > 1 {
				stats.MonthlyStats = stats.MonthlyStats[:1]
			}
		case "week":
			if len(stats.MonthlyStats) > 1 {
				stats.MonthlyStats = stats.MonthlyStats[:1]
			}
		}
	}

	if req.IncludeGoals {
		if err := s.repo.UpdateReadingGoalProgress(ctx, userID); err == nil {
			goals, err := s.repo.GetActiveReadingGoals(ctx, userID)
			if err == nil {
				stats.Goals = goals
			}
		}
	}

	return stats, nil
}

// GetReadingSummary returns lean reading statistics that are safe for empty users.
func (s *Service) GetReadingSummary(ctx context.Context, userID int64) (*ReadingSummary, error) {
	summary, err := s.repo.GetReadingSummary(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ReadingSummary{}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if summary == nil {
		return &ReadingSummary{}, nil
	}
	return summary, nil
}

// GetReadingAnalyticsBuckets returns grouped analytics and always succeeds with defaults.
func (s *Service) GetReadingAnalyticsBuckets(ctx context.Context, userID int64) (*ReadingAnalyticsResponse, error) {
	resp, err := s.repo.GetReadingAnalyticsBuckets(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ReadingAnalyticsResponse{
				Daily:   []ReadingAnalyticsPoint{},
				Weekly:  []ReadingAnalyticsPoint{},
				Monthly: []ReadingAnalyticsPoint{},
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if resp == nil {
		return &ReadingAnalyticsResponse{
			Daily:   []ReadingAnalyticsPoint{},
			Weekly:  []ReadingAnalyticsPoint{},
			Monthly: []ReadingAnalyticsPoint{},
		}, nil
	}
	return resp, nil
}
