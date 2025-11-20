package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/domain/history"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
)

var (
	ErrMangaNotFound         = errors.New("manga not found")
	ErrMangaAlreadyInLibrary = errors.New("manga already in library")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrDatabaseError         = errors.New("database error")
	ErrMangaNotInLibrary     = errors.New("manga not in library")
)

var validStatuses = map[string]bool{
	"plan_to_read": true,
	"reading":      true,
	"completed":    true,
	"on_hold":      true,
	"dropped":      true,
}

// ProgressProvider exposes progress retrieval
type ProgressProvider interface {
	GetProgress(ctx context.Context, userID, mangaID int64) (*history.UserProgress, error)
}

//go:generate mockgen -destination=./mocks/mock_progress_provider.go -package=library . ProgressProvider

// Service coordinates library use cases
type Service struct {
	repo         *Repository
	mangaService internalmanga.GetByID
	progressSvc  ProgressProvider
}

// NewService constructs library service
func NewService(repo *Repository, mangaService internalmanga.GetByID, progressSvc ProgressProvider) *Service {
	return &Service{repo: repo, mangaService: mangaService, progressSvc: progressSvc}
}

// AddToLibrary inserts manga into user's library
func (s *Service) AddToLibrary(ctx context.Context, userID, mangaID int64, req AddToLibraryRequest) (*AddToLibraryResponse, error) {
	if !validStatuses[req.Status] {
		return nil, ErrInvalidStatus
	}

	manga, err := s.mangaService.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if exists {
		return nil, ErrMangaAlreadyInLibrary
	}

	if err := s.repo.AddToLibrary(ctx, userID, mangaID, req.Status, req.CurrentChapter); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	libraryStatus, err := s.repo.GetLibraryStatus(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	var progress *history.UserProgress
	if s.progressSvc != nil {
		progress, _ = s.progressSvc.GetProgress(ctx, userID, mangaID)
	}

	return &AddToLibraryResponse{
		Message:       "manga added to library successfully",
		LibraryStatus: libraryStatus,
		UserProgress:  progress,
	}, nil
}

// RemoveFromLibrary deletes a manga from user's library
func (s *Service) RemoveFromLibrary(ctx context.Context, userID, mangaID int64) error {
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return ErrMangaNotInLibrary
	}

	if err := s.repo.RemoveFromLibrary(ctx, userID, mangaID); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return nil
}

// GetLibrary returns the user's library entries
func (s *Service) GetLibrary(ctx context.Context, userID int64) (*GetLibraryResponse, error) {
	entries, err := s.repo.GetLibrary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return &GetLibraryResponse{Entries: entries}, nil
}

// UpdateLibraryStatus updates a manga's status for the user
func (s *Service) UpdateLibraryStatus(ctx context.Context, userID, mangaID int64, req UpdateLibraryStatusRequest) (*LibraryStatus, error) {
	if !validStatuses[req.Status] {
		return nil, ErrInvalidStatus
	}

	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return nil, ErrMangaNotInLibrary
	}

	status, err := s.repo.UpdateLibraryStatus(ctx, userID, mangaID, req.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMangaNotInLibrary
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return status, nil
}

// CheckLibraryExists is exposed for other domains that need membership validation
func (s *Service) CheckLibraryExists(ctx context.Context, userID, mangaID int64) (bool, error) {
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return exists, nil
}

// GetLibraryStatus exposes repository lookup for composition with other domains
func (s *Service) GetLibraryStatus(ctx context.Context, userID, mangaID int64) (*LibraryStatus, error) {
	status, err := s.repo.GetLibraryStatus(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return status, nil
}
