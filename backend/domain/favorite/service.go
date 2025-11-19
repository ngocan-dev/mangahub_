package favorite

import (
	"context"
	"errors"
	"fmt"
<<<<<<< HEAD:backend/domain/favorite/service.go

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/rating"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
=======
>>>>>>> 6cdd9d56804641720cdb3be51e63eaf36cd18f99:backend/cmd/domain/favorite/service.go
)

var (
	ErrDatabaseError     = errors.New("database error")
	ErrMangaNotInLibrary = errors.New("manga not in library")
	ErrAlreadyFavorite   = errors.New("manga already in favorites")
)

// LibraryChecker ensures the manga exists in the user's library
type LibraryChecker interface {
	CheckLibraryExists(ctx context.Context, userID, mangaID int64) (bool, error)
}

<<<<<<< HEAD:backend/domain/favorite/service.go
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
=======
// Service coordinates favorite use cases
type Service struct {
	repo           *Repository
	libraryChecker LibraryChecker
}

// NewService constructs favorite service
func NewService(repo *Repository, libraryChecker LibraryChecker) *Service {
	return &Service{repo: repo, libraryChecker: libraryChecker}
>>>>>>> 6cdd9d56804641720cdb3be51e63eaf36cd18f99:backend/cmd/domain/favorite/service.go
}

// AddFavorite marks a manga as favorite for the user
func (s *Service) AddFavorite(ctx context.Context, userID, mangaID int64) error {
	exists, err := s.libraryChecker.CheckLibraryExists(ctx, userID, mangaID)
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
<<<<<<< HEAD:backend/domain/favorite/service.go

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
=======
	return nil
}

// RemoveFavorite removes a manga from user's favorites
func (s *Service) RemoveFavorite(ctx context.Context, userID, mangaID int64) error {
	exists, err := s.libraryChecker.CheckLibraryExists(ctx, userID, mangaID)
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
>>>>>>> 6cdd9d56804641720cdb3be51e63eaf36cd18f99:backend/cmd/domain/favorite/service.go
