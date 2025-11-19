package comment

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository handles database operations for reviews
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new comment repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateReview inserts a review row atomically
func (r *Repository) CreateReview(ctx context.Context, userID, mangaID int64, rating int, content string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var existingCount int
	err = tx.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM Reviews
        WHERE User_Id = ? AND Novel_Id = ?
    `, userID, mangaID).Scan(&existingCount)
	if err != nil {
		return 0, err
	}
	if existingCount > 0 {
		return 0, fmt.Errorf("user has already reviewed this manga")
	}

	result, err := tx.ExecContext(ctx, `
        INSERT INTO Reviews (User_Id, Novel_Id, Rating, Content, Created_At, Updated_At)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `, userID, mangaID, rating, content)
	if err != nil {
		return 0, err
	}

	reviewID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return reviewID, nil
}

// GetReviewByUserAndManga returns review if exists
func (r *Repository) GetReviewByUserAndManga(ctx context.Context, userID, mangaID int64) (*Review, error) {
	query := `
        SELECT r.Review_Id, r.User_Id, u.Username, r.Novel_Id, r.Rating, r.Content, r.Created_At, r.Updated_At
        FROM Reviews r
        JOIN Users u ON r.User_Id = u.UserId
        WHERE r.User_Id = ? AND r.Novel_Id = ?
    `
	var review Review
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&review.ReviewID,
		&review.UserID,
		&review.Username,
		&review.MangaID,
		&review.Rating,
		&review.Content,
		&review.CreatedAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		review.UpdatedAt = updatedAt.Time
	}

	return &review, nil
}

// GetReviews fetches paginated list of reviews
func (r *Repository) GetReviews(ctx context.Context, mangaID int64, page, limit int, sortBy string) ([]Review, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM Reviews WHERE Novel_Id = ?
    `, mangaID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Review{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if page > 10000 {
		page = 10000
	}
	offset := (page - 1) * limit

	orderClause := "ORDER BY r.Created_At DESC"
	if sortBy == "helpfulness" || sortBy == "rating" {
		orderClause = "ORDER BY r.Rating DESC, r.Created_At DESC"
	} else if sortBy == "oldest" {
		orderClause = "ORDER BY r.Created_At ASC"
	}

	query := fmt.Sprintf(`
        SELECT
            r.Review_Id,
            r.User_Id,
            u.Username,
            r.Novel_Id,
            r.Rating,
            r.Content,
            r.Created_At,
            r.Updated_At
        FROM Reviews r
        JOIN Users u ON r.User_Id = u.UserId
        WHERE r.Novel_Id = ?
        %s
        LIMIT ? OFFSET ?
    `, orderClause)

	rows, err := r.db.QueryContext(ctx, query, mangaID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&review.ReviewID,
			&review.UserID,
			&review.Username,
			&review.MangaID,
			&review.Rating,
			&review.Content,
			&review.CreatedAt,
			&updatedAt,
		); err != nil {
			continue
		}
		if updatedAt.Valid {
			review.UpdatedAt = updatedAt.Time
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// CheckMangaInCompletedLibrary ensures manga is completed by user
func (r *Repository) CheckMangaInCompletedLibrary(ctx context.Context, userID, mangaID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM User_Library
        WHERE User_Id = ? AND Novel_Id = ? AND Status = 'completed'
    `, userID, mangaID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetReviewStats aggregates review information
func (r *Repository) GetReviewStats(ctx context.Context, mangaID int64) (*ReviewStats, error) {
	query := `
        SELECT
            COUNT(*) as total_reviews,
            COALESCE(AVG(Rating), 0) as average_rating
        FROM Reviews
        WHERE Novel_Id = ?
    `
	var stats ReviewStats
	err := r.db.QueryRowContext(ctx, query, mangaID).Scan(
		&stats.TotalReviews,
		&stats.AverageRating,
	)
	if err != nil {
		return nil, err
	}
	stats.AverageRating = float64(int(stats.AverageRating*100+0.5)) / 100
	return &stats, nil
}
