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
