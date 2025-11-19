package favorite

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
)

var (
	ErrMangaNotFound         = errors.New("manga not found")
	ErrMangaAlreadyInLibrary = errors.New("manga already in library")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrDatabaseError         = errors.New("database error")
	ErrMangaNotInLibrary     = errors.New("manga not in library")
	ErrAlreadyFavorite       = errors.New("manga already in favorites")
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

// Service coordinates library use cases
type Service struct {
	repo         *Repository
	mangaService internalmanga.GetByID
	progressSvc  ProgressProvider
}

// NewService constructs favorite service
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

// AddFavorite marks a manga as favorite for the user
func (s *Service) AddFavorite(ctx context.Context, userID, mangaID int64) error {
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return ErrMangaNotInLibrary
	}

	isFavorite, err := s.repo.IsFavorite(ctx, userID, mangaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if isFavorite {
		return ErrAlreadyFavorite
	}

	if err := s.repo.AddFavorite(ctx, userID, mangaID); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return nil
}

// RemoveFavorite removes a manga from user's favorites
func (s *Service) RemoveFavorite(ctx context.Context, userID, mangaID int64) error {
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return ErrMangaNotInLibrary
	}

	isFavorite, err := s.repo.IsFavorite(ctx, userID, mangaID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !isFavorite {
		return nil
	}

	if err := s.repo.RemoveFavorite(ctx, userID, mangaID); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return nil
}

// GetFavorites returns all favorite entries for a user
func (s *Service) GetFavorites(ctx context.Context, userID int64) (*FavoritesResponse, error) {
	entries, err := s.repo.GetFavorites(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return &FavoritesResponse{Favorites: entries}, nil
}

// IsFavorite reports whether a manga is favorited by the user
func (s *Service) IsFavorite(ctx context.Context, userID, mangaID int64) (*FavoriteStatusResponse, error) {
	isFavorite, err := s.repo.IsFavorite(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return &FavoriteStatusResponse{IsFavorite: isFavorite}, nil
}
