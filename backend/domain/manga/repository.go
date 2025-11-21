package manga

import (
	"context"
	"database/sql"
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
	var (
		conditions []string
		args       []interface{}
	)

	useFullText := strings.TrimSpace(req.Query) != ""
	ftsQuery := buildFTSQuery(req.Query)

	if useFullText {
		args = append(args, ftsQuery)
	}

	// --- Filters ---
	if len(req.Genres) > 0 {
		placeholders := make([]string, len(req.Genres))
		for i, g := range req.Genres {
			placeholders[i] = "?"
			args = append(args, g)
		}
		conditions = append(conditions, "n.Genre IN ("+strings.Join(placeholders, ",")+")")
	}

	if req.Status != "" {
		conditions = append(conditions, "n.Status = ?")
		args = append(args, req.Status)
	}

	if req.MinRating != nil {
		conditions = append(conditions, "n.Rating_Point >= ?")
		args = append(args, *req.MinRating)
	}
	if req.MaxRating != nil {
		conditions = append(conditions, "n.Rating_Point <= ?")
		args = append(args, *req.MaxRating)
	}

	if req.YearFrom != nil {
		conditions = append(conditions, "n.Year >= ?")
		args = append(args, *req.YearFrom)
	}
	if req.YearTo != nil {
		conditions = append(conditions, "n.Year <= ?")
		args = append(args, *req.YearTo)
	}

	// --- Base query and relevance ranking ---
	baseQuery := `
        SELECT
            n.Novel_Id,
            n.Novel_Name,
            n.Title,
			n.Author,
            n.Genre,
            n.Status,
            n.Description,
            n.Image,
            n.Rating_Point,
            COALESCE(r.rank, 0) AS relevance
        FROM Novels n
    `

	countQuery := "SELECT COUNT(*) FROM Novels n"
	if useFullText {
		rankedCTE := `WITH ranked AS (
        SELECT rowid AS Novel_Id, bm25(NovelSearch, 1.0, 0.75) AS rank
        FROM NovelSearch
        WHERE NovelSearch MATCH ?
    )`
		baseQuery = rankedCTE + `
        SELECT
            n.Novel_Id,
            n.Novel_Name,
            n.Title,
            n.Author,
            n.Genre,
            n.Status,
            n.Description,
            n.Image,
            n.Rating_Point,
            r.rank AS relevance
        FROM Novels n
        JOIN ranked r ON n.Novel_Id = r.Novel_Id
    `
		countQuery = rankedCTE + `
        SELECT COUNT(*) FROM Novels n
        JOIN ranked r ON n.Novel_Id = r.Novel_Id
    `
	}

	// --- Build WHERE ---
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	// --- Sorting ---
	switch req.SortBy {
	case "rating":
		baseQuery += " ORDER BY n.Rating_Point DESC"
	case "date_updated":
		baseQuery += " ORDER BY n.Date_Updated DESC"
	case "relevance":
		baseQuery += " ORDER BY relevance ASC, n.Rating_Point DESC"
	default:
		if useFullText {
			baseQuery += " ORDER BY relevance ASC, n.Rating_Point DESC"
		} else {
			baseQuery += " ORDER BY n.Rating_Point DESC"
		}
	}

	// --- Pagination ---
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

	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// --- Execute Search Query ---
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
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
			&m.RelevanceScore,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// --- Count Query ---
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func buildFTSQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	parts := strings.Fields(query)
	for i, part := range parts {
		parts[i] = part + "*"
	}

	return strings.Join(parts, " ")
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

// GetPopularManga returns the most popular manga based on rating points
func (r *Repository) GetPopularManga(ctx context.Context, limit int) ([]Manga, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
        SELECT Novel_Id, Novel_Name, Title, Author, Genre, Status, Description, Image, Rating_Point
        FROM Novels
        ORDER BY Rating_Point DESC, Date_Updated DESC
        LIMIT ?
    `

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var popular []Manga
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
		); err != nil {
			return nil, err
		}
		popular = append(popular, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return popular, nil
}
