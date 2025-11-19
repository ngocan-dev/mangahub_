package favorite

import (
	"context"
	"errors"
	"fmt"
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

// Service coordinates favorite use cases
type Service struct {
	repo           *Repository
	libraryChecker LibraryChecker
}

// NewService constructs favorite service
func NewService(repo *Repository, libraryChecker LibraryChecker) *Service {
	return &Service{repo: repo, libraryChecker: libraryChecker}
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
