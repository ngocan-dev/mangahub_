package comment

import "time"

// Review represents a manga review
type Review struct {
	ReviewID  int64     `json:"review_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	MangaID   int64     `json:"manga_id"`
	Rating    int       `json:"rating"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ReviewStats contains aggregated review metrics
type ReviewStats struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
}

// CreateReviewRequest captures incoming payload
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=10"`
	Content string `json:"content" binding:"required,min=10,max=5000"`
}

// CreateReviewResponse is returned after successful creation
type CreateReviewResponse struct {
	Message string  `json:"message"`
	Review  *Review `json:"review"`
}

// UpdateReviewRequest captures partial review updates
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Content *string `json:"content"`
}

// UpdateReviewResponse represents an updated review payload
type UpdateReviewResponse struct {
	Message string  `json:"message"`
	Review  *Review `json:"review"`
}

// GetReviewsResponse represents paginated review listing
type GetReviewsResponse struct {
	Data []Review    `json:"data"`
	Meta ReviewsMeta `json:"meta"`
}

// ReviewsMeta contains pagination information
type ReviewsMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// DeleteReviewResponse acknowledges review deletion
type DeleteReviewResponse struct {
	Message string `json:"message"`
}
