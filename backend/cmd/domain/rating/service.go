package rating

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/security"
)

var (
	// ErrInvalidRating indicates the provided rating is outside acceptable range.
	ErrInvalidRating = errors.New("rating must be between 1 and 10")
	// ErrRepository signals persistence failures.
	ErrRepository = errors.New("rating repository error")
)

// UserRatingService exposes rating operations used by other domains.
type UserRatingService interface {
	SetRating(ctx context.Context, userID, mangaID int64, rating int) (*UserRating, error)
	GetUserRating(ctx context.Context, userID, mangaID int64) (*UserRating, error)
	GetAggregateRating(ctx context.Context, mangaID int64) (*AggregateRating, error)
}

// Service implements rating use cases.
type Service struct {
	repo Store
}

// Ensure Service implements UserRatingService.
var _ UserRatingService = (*Service)(nil)

// NewService constructs a rating service.
func NewService(repo Store) *Service {
	return &Service{repo: repo}
}

// SetRating validates and persists a user's rating.
func (s *Service) SetRating(ctx context.Context, userID, mangaID int64, rating int) (*UserRating, error) {
	if err := security.ValidateReviewRating(rating); err != nil {
		return nil, ErrInvalidRating
	}

	if err := s.repo.UpsertRating(ctx, userID, mangaID, rating); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepository, err)
	}

	// Update aggregate rating for manga.
	if agg, err := s.repo.GetAggregateRating(ctx, mangaID); err == nil && agg != nil {
		_ = s.repo.UpdateMangaAggregate(ctx, mangaID, agg.Average)
	}

	updated, err := s.repo.GetUserRating(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepository, err)
	}
	return updated, nil
}

// GetUserRating returns the user's rating if present.
func (s *Service) GetUserRating(ctx context.Context, userID, mangaID int64) (*UserRating, error) {
	rating, err := s.repo.GetUserRating(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepository, err)
	}
	return rating, nil
}

// GetAggregateRating returns aggregate rating stats for a manga.
func (s *Service) GetAggregateRating(ctx context.Context, mangaID int64) (*AggregateRating, error) {
	agg, err := s.repo.GetAggregateRating(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRepository, err)
	}
	return agg, nil
}
