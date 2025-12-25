package library

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	domainlibrary "github.com/ngocan-dev/mangahub/backend/domain/library"
)

// Repository handles persistence for library operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a library repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CheckLibraryExists ensures manga already in user's library
func (r *Repository) CheckLibraryExists(ctx context.Context, userID, mangaID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM user_library WHERE user_id = ? AND manga_id = ?`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddToLibrary inserts both library entry and initial progress transactionally.
// It returns true when the manga already exists in the user's library.
func (r *Repository) AddToLibrary(ctx context.Context, userID, mangaID int64, status string, currentChapter int) (bool, error) {
	now := time.Now()
	if _, err := r.db.ExecContext(ctx, `
INSERT INTO user_library (user_id, manga_id, status, current_chapter, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
`, userID, mangaID, status, currentChapter, now, now); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return true, nil
		}
		return false, err
	}

	var chapterID *int64
	if currentChapter > 0 {
		var chapterIDVal sql.NullInt64
		if err := r.db.QueryRowContext(ctx, `
SELECT id FROM chapters WHERE manga_id = ? AND number = ? LIMIT 1
`, mangaID, currentChapter).Scan(&chapterIDVal); err == nil && chapterIDVal.Valid {
			chapterID = &chapterIDVal.Int64
		}
	}

	if _, err := r.db.ExecContext(ctx, `
INSERT INTO reading_progress (user_id, manga_id, current_chapter_id, last_read_at, progress_percent, current_page)
VALUES (?, ?, ?, ?, 0, 0)
ON DUPLICATE KEY UPDATE current_chapter_id = VALUES(current_chapter_id), last_read_at = VALUES(last_read_at)
`, userID, mangaID, chapterID, now); err != nil {
		return false, err
	}

	return false, nil
}

// GetLibraryStatus fetches user's library status for manga
func (r *Repository) GetLibraryStatus(ctx context.Context, userID, mangaID int64) (*domainlibrary.LibraryStatus, error) {
	query := `
SELECT status, current_chapter, created_at, updated_at
FROM user_library
WHERE user_id = ? AND manga_id = ?
`
	var status domainlibrary.LibraryStatus
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&status.Status,
		&status.CurrentChapter,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		status.StartedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		status.CompletedAt = &updatedAt.Time
	}

	return &status, nil
}

// RemoveFromLibrary deletes library entry and related progress transactionally
func (r *Repository) RemoveFromLibrary(ctx context.Context, userID, mangaID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM reading_progress WHERE user_id = ? AND manga_id = ?`, userID, mangaID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM user_library WHERE user_id = ? AND manga_id = ?`, userID, mangaID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// UpdateLibraryStatus updates status timestamps and returns new status
func (r *Repository) UpdateLibraryStatus(ctx context.Context, userID, mangaID int64, status string) (*domainlibrary.LibraryStatus, error) {
	result, err := r.db.ExecContext(ctx, `
UPDATE user_library
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND manga_id = ?
`, status, userID, mangaID)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetLibraryStatus(ctx, userID, mangaID)
}

// GetLibrary fetches the user's library listing
func (r *Repository) GetLibrary(ctx context.Context, userID int64) ([]domainlibrary.LibraryEntry, error) {
	query := `
SELECT ul.manga_id, m.title, m.cover_url, ul.status, ul.current_chapter, ul.created_at, ul.updated_at
FROM user_library ul
JOIN mangas m ON m.id = ul.manga_id
WHERE ul.user_id = ?
ORDER BY ul.updated_at DESC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domainlibrary.LibraryEntry
	for rows.Next() {
		var entry domainlibrary.LibraryEntry
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&entry.MangaID, &entry.Title, &entry.CoverImage, &entry.Status, &entry.CurrentChapter, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			entry.StartedAt = &createdAt.Time
		}
		if updatedAt.Valid {
			entry.CompletedAt = &updatedAt.Time
			entry.LastUpdated = updatedAt.Time
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
