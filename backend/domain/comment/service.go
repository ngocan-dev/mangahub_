package comment

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/domain/rating"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/security"
)

var (
	ErrMangaNotFound         = errors.New("manga not found")
	ErrMangaNotCompleted     = errors.New("manga must be in completed list to write review")
	ErrReviewAlreadyExists   = errors.New("review already exists for this manga")
	ErrInvalidReviewRating   = errors.New("rating must be between 1 and 10")
	ErrReviewContentTooShort = errors.New("review content must be at least 10 characters")
	ErrReviewContentTooLong  = errors.New("review content must not exceed 5000 characters")
	ErrDatabaseError         = errors.New("database error")
)

// RatingSetter abstracts rating updates
type RatingSetter interface {
	SetRating(ctx context.Context, userID, mangaID int64, rating int) (*rating.UserRating, error)
}

// Service handles review use cases
type Service struct {
	repo          *Repository
	mangaService  internalmanga.GetByID
	ratingService RatingSetter
}

// NewService builds a review service
func NewService(repo *Repository, mangaService internalmanga.GetByID, ratingService RatingSetter) *Service {
	return &Service{repo: repo, mangaService: mangaService, ratingService: ratingService}
}

// CreateReview stores a new review after validations
func (s *Service) CreateReview(ctx context.Context, userID, mangaID int64, req CreateReviewRequest) (*CreateReviewResponse, error) {
	if err := security.ValidateReviewRating(req.Rating); err != nil {
		return nil, ErrInvalidReviewRating
	}

	if err := security.ValidateReviewContent(req.Content); err != nil {
		switch {
		case errors.Is(err, security.ErrInputTooShort):
			return nil, ErrReviewContentTooShort
		case errors.Is(err, security.ErrInputTooLong):
			return nil, ErrReviewContentTooLong
		case errors.Is(err, security.ErrContainsSQLInjection):
			return nil, fmt.Errorf("invalid input: %w", err)
		default:
			return nil, fmt.Errorf("validation error: %w", err)
		}
	}

	sanitized := security.SanitizeReviewContent(req.Content)

	manga, err := s.mangaService.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	completed, err := s.repo.CheckMangaInCompletedLibrary(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !completed {
		return nil, ErrMangaNotCompleted
	}

	existing, err := s.repo.GetReviewByUserAndManga(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if existing != nil {
		return nil, ErrReviewAlreadyExists
	}

	if _, err := s.repo.CreateReview(ctx, userID, mangaID, req.Rating, sanitized); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if s.ratingService != nil {
		if _, err := s.ratingService.SetRating(ctx, userID, mangaID, req.Rating); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
	}
	review, err := s.repo.GetReviewByUserAndManga(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &CreateReviewResponse{
		Message: "review created successfully",
		Review:  review,
	}, nil
}

// GetReviews returns paginated review list
func (s *Service) GetReviews(ctx context.Context, mangaID int64, page, limit int, sortBy string) (*GetReviewsResponse, error) {
	reviews, total, err := s.repo.GetReviews(ctx, mangaID, page, limit, sortBy)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	stats, err := s.repo.GetReviewStats(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	pages := 0
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}

	return &GetReviewsResponse{
		Reviews:       reviews,
		Total:         total,
		Page:          page,
		Limit:         limit,
		Pages:         pages,
		AverageRating: stats.AverageRating,
		TotalReviews:  stats.TotalReviews,
	}, nil
}
