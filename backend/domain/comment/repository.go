package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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

// CheckMangaInCompletedLibrary verifies if the manga exists in the user's completed list
func (r *Repository) CheckMangaInCompletedLibrary(ctx context.Context, userID, mangaID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM user_library
        WHERE user_id = ? AND manga_id = ? AND status = 'completed'
    `, userID, mangaID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

// GetReviewByID fetches a review by its identifier
func (r *Repository) GetReviewByID(ctx context.Context, reviewID int64) (*Review, error) {
	query := `
SELECT r.Review_Id, r.User_Id, u.Username, r.Novel_Id, r.Rating, r.Content, r.Created_At, r.Updated_At
FROM Reviews r
JOIN Users u ON r.User_Id = u.UserId
WHERE r.Review_Id = ?
`
	var review Review
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, reviewID).Scan(
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

// GetReviewsByMangaID fetches paginated list of reviews for a manga.
func (r *Repository) GetReviewsByMangaID(ctx context.Context, mangaID int64, page, limit int, sortBy string) ([]Review, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ratings WHERE manga_id = ? AND review IS NOT NULL AND review <> ''`, mangaID).Scan(&total)
	if err != nil {
		if isNoDataError(err) {
			return []Review{}, 0, nil
		}
		log.Printf("comment.repository.GetReviewsByMangaID: count failed manga_id=%d err=%v", mangaID, err)
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
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	orderClause := "ORDER BY r.created_at DESC"
	switch strings.ToLower(sortBy) {
	case "rating", "helpfulness":
		orderClause = "ORDER BY r.score DESC, r.created_at DESC"
	case "oldest":
		orderClause = "ORDER BY r.created_at ASC"
	}

	query := fmt.Sprintf(`
        SELECT
            r.id,
            r.user_id,
            u.username,
            u.avatar_url,
            r.manga_id,
            r.score,
            r.review,
            r.created_at,
            r.updated_at
        FROM ratings r
        LEFT JOIN users u ON r.user_id = u.id
        WHERE r.manga_id = ? AND r.review IS NOT NULL AND r.review <> ''
        %s
        LIMIT ? OFFSET ?
    `, orderClause)

	rows, err := r.db.QueryContext(ctx, query, mangaID, limit, offset)
	if err != nil {
		log.Printf("comment.repository.GetReviewsByMangaID: query failed manga_id=%d err=%v", mangaID, err)
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var username sql.NullString
		var avatar sql.NullString
		var content sql.NullString
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&review.ReviewID,
			&review.UserID,
			&username,
			&avatar,
			&review.MangaID,
			&review.Rating,
			&content,
			&review.CreatedAt,
			&updatedAt,
		); err != nil {
			log.Printf("comment.repository.GetReviewsByMangaID: scan failed manga_id=%d err=%v", mangaID, err)
			return nil, 0, err
		}

		if username.Valid {
			review.Username = username.String
		}
		if avatar.Valid {
			review.AvatarURL = avatar.String
		}
		if content.Valid {
			review.Content = content.String
		}
		if updatedAt.Valid {
			review.UpdatedAt = updatedAt.Time
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		log.Printf("comment.repository.GetReviewsByMangaID: rows error manga_id=%d err=%v", mangaID, err)
		return nil, 0, err
	}

	return reviews, total, nil
}

// GetReviewStats aggregates review information
func (r *Repository) GetReviewStats(ctx context.Context, mangaID int64) (*ReviewStats, error) {
	query := `
        SELECT
            COALESCE(COUNT(*), 0) as total_reviews,
            COALESCE(AVG(score), 0) as average_rating
        FROM ratings
        WHERE manga_id = ? AND review IS NOT NULL AND review <> ''
    `
	var stats ReviewStats
	err := r.db.QueryRowContext(ctx, query, mangaID).Scan(
		&stats.TotalReviews,
		&stats.AverageRating,
	)
	if err != nil {
		if isNoDataError(err) {
			return &ReviewStats{}, nil
		}
		log.Printf("comment.repository.GetReviewStats: query failed manga_id=%d err=%v", mangaID, err)
		return nil, err
	}
	stats.AverageRating = float64(int(stats.AverageRating*100+0.5)) / 100
	return &stats, nil
}

func isNoDataError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "no such table")
}

// UpdateReview applies provided field changes on a review
func (r *Repository) UpdateReview(ctx context.Context, reviewID, userID int64, rating *int, content *string) error {
	setClauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 4)
	if rating != nil {
		setClauses = append(setClauses, "Rating = ?")
		args = append(args, *rating)
	}
	if content != nil {
		setClauses = append(setClauses, "Content = ?")
		args = append(args, *content)
	}
	if len(setClauses) == 0 {
		return nil
	}
	setClauses = append(setClauses, "Updated_At = CURRENT_TIMESTAMP")
	query := fmt.Sprintf(`UPDATE Reviews SET %s WHERE Review_Id = ? AND User_Id = ?`, strings.Join(setClauses, ", "))
	args = append(args, reviewID, userID)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteReview removes a review owned by the user
func (r *Repository) DeleteReview(ctx context.Context, reviewID, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM Reviews WHERE Review_Id = ? AND User_Id = ?`, reviewID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
