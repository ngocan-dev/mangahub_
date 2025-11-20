package rating

import (
	"context"
	"database/sql"
)

// Store defines the persistence operations needed by the rating service.
type Store interface {
	UpsertRating(ctx context.Context, userID, mangaID int64, rating int) error
	GetUserRating(ctx context.Context, userID, mangaID int64) (*UserRating, error)
	GetAggregateRating(ctx context.Context, mangaID int64) (*AggregateRating, error)
	UpdateMangaAggregate(ctx context.Context, mangaID int64, average float64) error
}

// Repository persists rating data in SQL database.
type Repository struct {
	db *sql.DB
}

// Ensure Repository implements Store interface.
var _ Store = (*Repository)(nil)

// NewRepository constructs a SQL-backed rating repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UpsertRating inserts or updates a user's rating for a manga.
func (r *Repository) UpsertRating(ctx context.Context, userID, mangaID int64, rating int) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO Rating_System (User_Id, Novel_Id, Rating_Value, Rating_Date)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(User_Id, Novel_Id) DO UPDATE SET
            Rating_Value = excluded.Rating_Value,
            Rating_Date = CURRENT_TIMESTAMP
    `, userID, mangaID, rating)
	return err
}

// GetUserRating fetches a user's rating for a manga if it exists.
func (r *Repository) GetUserRating(ctx context.Context, userID, mangaID int64) (*UserRating, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT Rating_Id, User_Id, Novel_Id, Rating_Value, Rating_Date
        FROM Rating_System
        WHERE User_Id = ? AND Novel_Id = ?
    `, userID, mangaID)

	var rating UserRating
	if err := row.Scan(&rating.ID, &rating.UserID, &rating.MangaID, &rating.Rating, &rating.RatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}

// GetAggregateRating returns aggregate rating statistics for a manga.
func (r *Repository) GetAggregateRating(ctx context.Context, mangaID int64) (*AggregateRating, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT ?, COALESCE(AVG(Rating_Value), 0), COUNT(Rating_Value)
FROM Rating_System
WHERE Novel_Id = ?
`, mangaID, mangaID)

	var agg AggregateRating
	if err := row.Scan(&agg.MangaID, &agg.Average, &agg.Count); err != nil {
		return nil, err
	}
	return &agg, nil
}

// UpdateMangaAggregate writes the aggregate rating back to the manga record.
func (r *Repository) UpdateMangaAggregate(ctx context.Context, mangaID int64, average float64) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE Novels SET Rating_Point = ? WHERE Novel_Id = ?
    `, average, mangaID)
	return err
}
