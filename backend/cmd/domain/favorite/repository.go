package favorite

import (
	"context"
	"database/sql"
	"time"
)

// Repository handles persistence for library operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a favorite repository
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
func (r *Repository) AddToLibrary(ctx context.Context, userID, mangaID int64, status string, currentChapter int, isFavorite bool) error {
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

	favInt := 0
	if isFavorite {
		favInt = 1
	}

	if _, err := tx.ExecContext(ctx, `
        INSERT INTO User_Library (User_Id, Novel_Id, Status, Is_Favorite, Started_At, Completed_At, Last_Updated_At)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, userID, mangaID, status, favInt, startedAt, completedAt, now); err != nil {
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
func (r *Repository) GetLibraryStatus(ctx context.Context, userID, mangaID int64) (*LibraryStatus, error) {
	query := `
	SELECT Status, Is_Favorite, Started_At, Completed_At
	FROM User_Library
	WHERE User_Id = ? AND Novel_Id = ?
`
	var status LibraryStatus
	var favInt int
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&status.Status,
		&favInt,
		&startedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	status.IsFavorite = favInt != 0
	if startedAt.Valid {
		status.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}

	return &status, nil
}
