package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/security"
	internalhistory "github.com/ngocan-dev/mangahub/manga-backend/internal/history"
	internalmanga "github.com/ngocan-dev/mangahub/manga-backend/internal/manga"
)

var (
	ErrMangaNotFound         = errors.New("manga not found")
	ErrMangaNotCompleted     = errors.New("manga must be in completed list to write review")
	ErrReviewAlreadyExists   = errors.New("review already exists for this manga")
	ErrInvalidReviewRating   = errors.New("rating must be between 1 and 10")
	ErrReviewContentTooShort = errors.New("review content must be at least 10 characters")
	ErrReviewContentTooLong  = errors.New("review content must not exceed 5000 characters")
	ErrReviewNotFound        = errors.New("review not found")
	ErrNoReviewChanges       = errors.New("no review fields provided for update")
	ErrDatabaseError         = errors.New("database error")
)

// Service handles review use cases
type Service struct {
	repo           *Repository
	mangaService   internalmanga.GetByID
	historyService internalhistory.CompletionChecker
}

// NewService builds a review service
func NewService(repo *Repository, mangaService internalmanga.GetByID, historyService internalhistory.CompletionChecker) *Service {
	return &Service{repo: repo, mangaService: mangaService, historyService: historyService}
}

// CreateReview stores a new review after validations
func (s *Service) CreateReview(ctx context.Context, userID, mangaID int64, req CreateReviewRequest) (*CreateReviewResponse, error) {
	if err := security.ValidateReviewRating(req.Rating); err != nil {
		return nil, ErrInvalidReviewRating
	}

	if err := security.ValidateReviewContent(req.Content); err != nil {
		return nil, mapContentValidationError(err)
	}

	sanitized := security.SanitizeReviewContent(req.Content)

	manga, err := s.mangaService.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	completed, err := s.historyService.HasCompletedManga(ctx, userID, mangaID)
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

// GetReviewStats returns aggregated statistics for a manga
func (s *Service) GetReviewStats(ctx context.Context, mangaID int64) (*ReviewStats, error) {
	stats, err := s.repo.GetReviewStats(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return stats, nil
}

// UpdateReview updates rating/content of a review
func (s *Service) UpdateReview(ctx context.Context, userID, reviewID int64, req UpdateReviewRequest) (*UpdateReviewResponse, error) {
	var rating *int
	var content *string
	if req.Rating != nil {
		if err := security.ValidateReviewRating(*req.Rating); err != nil {
			return nil, ErrInvalidReviewRating
		}
		rating = req.Rating
	}
	if req.Content != nil {
		if err := security.ValidateReviewContent(*req.Content); err != nil {
			return nil, mapContentValidationError(err)
		}
		sanitized := security.SanitizeReviewContent(*req.Content)
		content = &sanitized
	}
	if rating == nil && content == nil {
		return nil, ErrNoReviewChanges
	}
	if err := s.repo.UpdateReview(ctx, reviewID, userID, rating, content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if review == nil {
		return nil, ErrReviewNotFound
	}
	return &UpdateReviewResponse{
		Message: "review updated successfully",
		Review:  review,
	}, nil
}

// DeleteReview removes a review owned by the user
func (s *Service) DeleteReview(ctx context.Context, userID, reviewID int64) (*DeleteReviewResponse, error) {
	if err := s.repo.DeleteReview(ctx, reviewID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return &DeleteReviewResponse{Message: "review deleted successfully"}, nil
}

func mapContentValidationError(err error) error {
	switch {
	case errors.Is(err, security.ErrInputTooShort):
		return ErrReviewContentTooShort
	case errors.Is(err, security.ErrInputTooLong):
		return ErrReviewContentTooLong
	case errors.Is(err, security.ErrContainsSQLInjection):
		return fmt.Errorf("invalid input: %w", err)
	default:
		return fmt.Errorf("validation error: %w", err)
	}
}
