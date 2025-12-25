package comment

import (
	"context"
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub/backend/domain/rating"
	"github.com/ngocan-dev/mangahub/backend/internal/security"
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

// MangaGetter retrieves manga information
type MangaGetter interface {
	Exists(ctx context.Context, mangaID int64) (bool, error)
}

// ActivityRecorder logs user activity to downstream feeds.
type ActivityRecorder interface {
	RecordActivity(ctx context.Context, userID int64, activityType string, mangaID *int64, payload map[string]interface{}) error
}

// Service handles review use cases
type Service struct {
	repo          *Repository
	mangaService  MangaGetter
	ratingService RatingSetter
	activityLog   ActivityRecorder
}

// NewService builds a review service
func NewService(repo *Repository, mangaService MangaGetter, ratingService RatingSetter) *Service {
	return &Service{repo: repo, mangaService: mangaService, ratingService: ratingService}
}

// SetActivityRecorder configures the optional activity recorder
func (s *Service) SetActivityRecorder(recorder ActivityRecorder) {
	s.activityLog = recorder
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

	mangaExists, err := s.mangaService.Exists(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !mangaExists {
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
	if s.activityLog != nil {
		_ = s.activityLog.RecordActivity(ctx, userID, "REVIEW", &mangaID, map[string]interface{}{
			"rating":  req.Rating,
			"content": sanitized,
		})
	}

	return &CreateReviewResponse{
		Message: "review created successfully",
		Review:  review,
	}, nil
}

// GetReviews returns paginated review list
func (s *Service) GetReviews(ctx context.Context, mangaID int64, page, limit int, sortBy string) (*GetReviewsResponse, error) {
	page, limit = normalizePagination(page, limit)

	reviews, total, err := s.repo.GetReviewsByMangaID(ctx, mangaID, page, limit, sortBy)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if reviews == nil {
		reviews = []Review{}
	}

	return &GetReviewsResponse{
		Data: reviews,
		Meta: ReviewsMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
