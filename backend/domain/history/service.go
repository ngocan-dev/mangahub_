package history

import (
	"context"
	"errors"
	"fmt"
	"time"

	internalchapter "github.com/ngocan-dev/mangahub/manga-backend/internal/chapter"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
)

var (
	ErrInvalidChapterNumber = errors.New("invalid chapter number")
	ErrMangaNotFound        = errors.New("manga not found")
	ErrMangaNotInLibrary    = errors.New("manga not in library")
	ErrDatabaseError        = errors.New("database error")
)

// ChapterService exposes chapter operations needed by history
type ChapterService interface {
	ValidateChapter(ctx context.Context, mangaID int64, chapter int) (*internalchapter.ChapterSummary, error)
}

// LibraryChecker verifies library membership
type LibraryChecker interface {
	CheckLibraryExists(ctx context.Context, userID, mangaID int64) (bool, error)
}

// Broadcaster broadcasts progress updates
type Broadcaster interface {
	BroadcastProgress(ctx context.Context, userID, mangaID int64, chapter int, chapterID *int64) error
}

// Service manages history use cases
type Service struct {
	repo           *Repository
	chapterService ChapterService
	libraryChecker LibraryChecker
	broadcaster    Broadcaster
	mangaService   internalmanga.GetByID
}

// NewService builds history service
func NewService(repo *Repository, chapterSvc ChapterService, libraryChecker LibraryChecker, mangaService internalmanga.GetByID) *Service {
	return &Service{
		repo:           repo,
		chapterService: chapterSvc,
		libraryChecker: libraryChecker,
		mangaService:   mangaService,
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

	manga, err := s.mangaService.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	exists, err := s.libraryChecker.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
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
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	pages := 0
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}
	return &ActivityFeedResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		Limit:      limit,
		Pages:      pages,
	}, nil
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

	if err := s.repo.SaveReadingStatistics(ctx, stats); err != nil {
		// ignore cache error
	}

	return stats, nil
}

// GetReadingAnalytics filters stats
func (s *Service) GetReadingAnalytics(ctx context.Context, userID int64, req ReadingAnalyticsRequest) (*ReadingStatistics, error) {
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
