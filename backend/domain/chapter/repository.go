package chapter

import (
	"context"
	"database/sql"
)

// Repository encapsulates chapter queries
type Repository struct {
	db *sql.DB
}

// NewRepository builds repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetChapterCount returns total chapter count
func (r *Repository) GetChapterCount(ctx context.Context, mangaID int64) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM Chapters WHERE Novel_Id = ?`, mangaID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ValidateChapterNumber ensures chapter exists
func (r *Repository) ValidateChapterNumber(ctx context.Context, mangaID int64, chapterNumber int) (bool, *int64, error) {
	if chapterNumber < 1 {
		return false, nil, nil
	}
	query := `SELECT Chapter_Id FROM Chapters WHERE Novel_Id = ? AND Chapter_Number = ? LIMIT 1`
	var chapterID int64
	err := r.db.QueryRowContext(ctx, query, mangaID, chapterNumber).Scan(&chapterID)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return true, &chapterID, nil
}

// GetMaxChapterNumber returns highest chapter number
func (r *Repository) GetMaxChapterNumber(ctx context.Context, mangaID int64) (int, error) {
	var maxChapter sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT MAX(Chapter_Number) FROM Chapters WHERE Novel_Id = ?`, mangaID).Scan(&maxChapter); err != nil {
		return 0, err
	}
	if !maxChapter.Valid {
		return 0, nil
	}
	return int(maxChapter.Int64), nil
}
