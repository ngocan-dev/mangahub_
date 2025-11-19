package manga

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository handles manga metadata queries
type Repository struct {
	db *sql.DB
}

// NewRepository creates repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Search searches for manga/novels based on criteria
func (r *Repository) Search(ctx context.Context, req SearchRequest) ([]Manga, int, error) {
	query := `
        SELECT
            Novel_Id,
            Novel_Name,
            Title,
            Author,
            Genre,
            Status,
            Description,
            Image,
            Rating_Point
        FROM Novels
    `
	var conditions []string
	var args []interface{}

	if req.Query != "" {
		conditions = append(conditions, "(Novel_Name LIKE ? OR Title LIKE ? OR Author LIKE ? OR Description LIKE ?)")
		like := fmt.Sprintf("%%%s%%", req.Query)
		args = append(args, like, like, like, like)
	}
	if len(req.Genres) > 0 {
		placeholders := make([]string, len(req.Genres))
		for i, genre := range req.Genres {
			placeholders[i] = "?"
			args = append(args, genre)
		}
		conditions = append(conditions, fmt.Sprintf("Genre IN (%s)", strings.Join(placeholders, ",")))
	}
	if req.Status != "" {
		conditions = append(conditions, "Status = ?")
		args = append(args, req.Status)
	}
	if req.MinRating != nil {
		conditions = append(conditions, "Rating_Point >= ?")
		args = append(args, *req.MinRating)
	}
	if req.MaxRating != nil {
		conditions = append(conditions, "Rating_Point <= ?")
		args = append(args, *req.MaxRating)
	}
	if req.YearFrom != nil {
		conditions = append(conditions, "Year >= ?")
		args = append(args, *req.YearFrom)
	}
	if req.YearTo != nil {
		conditions = append(conditions, "Year <= ?")
		args = append(args, *req.YearTo)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch req.SortBy {
	case "rating":
		query += " ORDER BY Rating_Point DESC"
	case "date_updated":
		query += " ORDER BY Date_Updated DESC"
	default:
		query += " ORDER BY Rating_Point DESC"
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []Manga
	for rows.Next() {
		var m Manga
		if err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Title,
			&m.Author,
			&m.Genre,
			&m.Status,
			&m.Description,
			&m.Image,
			&m.RatingPoint,
		); err == nil {
			results = append(results, m)
		}
	}

	countQuery := "SELECT COUNT(*) FROM Novels"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	var total int
	countArgs := append([]interface{}{}, args[:len(args)-2]...)
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetByID retrieves manga details by ID
func (r *Repository) GetByID(ctx context.Context, mangaID int64) (*Manga, error) {
	query := `
        SELECT Novel_Id, Novel_Name, Title, Author, Genre, Status, Description, Image, Rating_Point
        FROM Novels
        WHERE Novel_Id = ?
    `
	var m Manga
	err := r.db.QueryRowContext(ctx, query, mangaID).Scan(
		&m.ID,
		&m.Name,
		&m.Title,
		&m.Author,
		&m.Genre,
		&m.Status,
		&m.Description,
		&m.Image,
		&m.RatingPoint,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}
