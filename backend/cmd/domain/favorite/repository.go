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

func (r *Repository) GetLibraryStatus(ctx context.Context, userID, mangaID int64) (*LibraryStatus, error) {
	query := `
SELECT Status, Rating, Started_At, Completed_At
FROM User_Library
WHERE User_Id = ? AND Novel_Id = ?
`
	var status LibraryStatus
	var rating sql.NullInt64
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&status.Status,
		&rating,
		&startedAt,
		&completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if rating.Valid {
		value := int(rating.Int64)
		status.Rating = &value
	}
	if startedAt.Valid {
		status.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}

	return &status, nil
}

// AddFavorite marks an existing library entry as favorite
func (r *Repository) AddFavorite(ctx context.Context, userID, mangaID int64) error {
	query := `
UPDATE User_Library
SET Is_Favorite = 1, Last_Updated_At = ?
WHERE User_Id = ? AND Novel_Id = ?
`
	result, err := r.db.ExecContext(ctx, query, time.Now(), userID, mangaID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RemoveFavorite unmarks a favorite entry
func (r *Repository) RemoveFavorite(ctx context.Context, userID, mangaID int64) error {
	query := `
UPDATE User_Library
SET Is_Favorite = 0, Last_Updated_At = ?
WHERE User_Id = ? AND Novel_Id = ?
`
	result, err := r.db.ExecContext(ctx, query, time.Now(), userID, mangaID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetFavorites returns all favorited manga for a user
func (r *Repository) GetFavorites(ctx context.Context, userID int64) ([]FavoriteEntry, error) {
	query := `
SELECT ul.Novel_Id, n.Title, n.Image, ul.Last_Updated_At
FROM User_Library ul
JOIN Novels n ON n.Novel_Id = ul.Novel_Id
WHERE ul.User_Id = ? AND ul.Is_Favorite = 1
ORDER BY ul.Last_Updated_At DESC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []FavoriteEntry
	for rows.Next() {
		var entry FavoriteEntry
		var addedAt sql.NullTime
		if err := rows.Scan(&entry.MangaID, &entry.Title, &entry.CoverImage, &addedAt); err != nil {
			return nil, err
		}
		if addedAt.Valid {
			entry.AddedAt = &addedAt.Time
		}
		favorites = append(favorites, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return favorites, nil
}

// IsFavorite checks if a manga is marked as favorite
func (r *Repository) IsFavorite(ctx context.Context, userID, mangaID int64) (bool, error) {
	query := `
SELECT Is_Favorite
FROM User_Library
WHERE User_Id = ? AND Novel_Id = ?
`
	var favInt int
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(&favInt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return favInt != 0, nil
}
