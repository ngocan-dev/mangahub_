package favorite

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/rating"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
)

var (
	ErrMangaNotFound         = errors.New("manga not found")
	ErrMangaAlreadyInLibrary = errors.New("manga already in library")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrDatabaseError         = errors.New("database error")
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

// RatingProvider exposes rating retrieval for library items.
type RatingProvider interface {
	GetUserRating(ctx context.Context, userID, mangaID int64) (*rating.UserRating, error)
}

// Service coordinates library use cases
type Service struct {
	repo         *Repository
	mangaService internalmanga.GetByID
	progressSvc  ProgressProvider
	ratingSvc    RatingProvider
}

// NewService constructs favorite service
func NewService(repo *Repository, mangaService internalmanga.GetByID, progressSvc ProgressProvider, ratingSvc RatingProvider) *Service {
	return &Service{repo: repo, mangaService: mangaService, progressSvc: progressSvc, ratingSvc: ratingSvc}
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

	if err := s.repo.AddToLibrary(ctx, userID, mangaID, req.Status, req.CurrentChapter, req.IsFavorite); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	libraryStatus, err := s.repo.GetLibraryStatus(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if s.ratingSvc != nil && libraryStatus != nil {
		if userRating, err := s.ratingSvc.GetUserRating(ctx, userID, mangaID); err == nil && userRating != nil {
			value := userRating.Rating
			libraryStatus.Rating = &value
		}
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
