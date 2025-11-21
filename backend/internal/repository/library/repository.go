package library

import (
	"context"
	"database/sql"
	"time"

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
	query := `SELECT COUNT(*) FROM User_Library WHERE User_Id = ? AND Novel_Id = ?`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddToLibrary inserts both library entry and initial progress transactionally
func (r *Repository) AddToLibrary(ctx context.Context, userID, mangaID int64, status string, currentChapter int) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	var startedAt *time.Time
	if status == "reading" || status == "completed" {
		startedAt = &now
	}
	var completedAt *time.Time
	if status == "completed" {
		completedAt = &now
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO User_Library (User_Id, Novel_Id, Status, Is_Favorite, Started_At, Completed_At, Last_Updated_At)
VALUES (?, ?, ?, 0, ?, ?, ?)
`, userID, mangaID, status, startedAt, completedAt, now); err != nil {
		return err
	}

	if currentChapter < 1 {
		currentChapter = 1
	}

	var chapterID *int64
	var chapterIDVal sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
        SELECT Chapter_Id FROM Chapters
        WHERE Novel_Id = ? AND Chapter_Number = ?
        LIMIT 1
    `, mangaID, currentChapter).Scan(&chapterIDVal); err == nil && chapterIDVal.Valid {
		chapterID = &chapterIDVal.Int64
	}

	if _, err := tx.ExecContext(ctx, `
        INSERT INTO Reading_Progress (User_Id, Novel_Id, Current_Chapter, Current_Chapter_Id, Last_Read_At)
        VALUES (?, ?, ?, ?, ?)
    `, userID, mangaID, currentChapter, chapterID, now); err != nil {
		return err
	}

	return tx.Commit()
}

// GetLibraryStatus fetches user's library status for manga
func (r *Repository) GetLibraryStatus(ctx context.Context, userID, mangaID int64) (*domainlibrary.LibraryStatus, error) {
	query := `
SELECT Status, Started_At, Completed_At
FROM User_Library
WHERE User_Id = ? AND Novel_Id = ?
`
	var status domainlibrary.LibraryStatus
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&status.Status,
		&startedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		status.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM Reading_Progress WHERE User_Id = ? AND Novel_Id = ?`, userID, mangaID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM User_Library WHERE User_Id = ? AND Novel_Id = ?`, userID, mangaID)
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
	now := time.Now()
	var startedAt, completedAt *time.Time
	if status == "reading" || status == "completed" {
		startedAt = &now
	}
	if status == "completed" {
		completedAt = &now
	}

	result, err := r.db.ExecContext(ctx, `
UPDATE User_Library
SET Status = ?, Started_At = COALESCE(Started_At, ?), Completed_At = CASE WHEN ? IS NOT NULL THEN ? ELSE Completed_At END, Last_Updated_At = ?
WHERE User_Id = ? AND Novel_Id = ?
`, status, startedAt, completedAt, completedAt, now, userID, mangaID)
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
SELECT ul.Novel_Id, n.Title, n.Image, ul.Status, ul.Started_At, ul.Completed_At, ul.Last_Updated_At
FROM User_Library ul
JOIN Novels n ON n.Novel_Id = ul.Novel_Id
WHERE ul.User_Id = ?
ORDER BY ul.Last_Updated_At DESC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domainlibrary.LibraryEntry
	for rows.Next() {
		var entry domainlibrary.LibraryEntry
		var startedAt, completedAt sql.NullTime
		if err := rows.Scan(&entry.MangaID, &entry.Title, &entry.CoverImage, &entry.Status, &startedAt, &completedAt, &entry.LastUpdated); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			entry.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			entry.CompletedAt = &completedAt.Time
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
