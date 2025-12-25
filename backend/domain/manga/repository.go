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
	var (
		conditions []string
		args       []interface{}
	)

	trimmedQuery := strings.TrimSpace(req.Query)
	if trimmedQuery != "" {
		like := "%" + trimmedQuery + "%"
		conditions = append(conditions, "(m.title LIKE ? OR m.synopsis LIKE ? OR m.alt_title LIKE ?)")
		args = append(args, like, like, like)
	}

	// --- Filters ---
	if len(req.Genres) > 0 {
		placeholders := make([]string, len(req.Genres))
		for i, g := range req.Genres {
			placeholders[i] = "?"
			args = append(args, g)
		}
		conditions = append(conditions, fmt.Sprintf(`
m.id IN (
    SELECT mt.manga_id
    FROM manga_tags mt
    JOIN tags t2 ON t2.id = mt.tag_id
    WHERE t2.name IN (%s)
    GROUP BY mt.manga_id
    HAVING COUNT(DISTINCT t2.name) = %d
)`, strings.Join(placeholders, ","), len(req.Genres)))
	}

	if req.Status != "" {
		conditions = append(conditions, "m.status = ?")
		args = append(args, req.Status)
	}

	if req.MinRating != nil {
		conditions = append(conditions, "m.rating_average >= ?")
		args = append(args, *req.MinRating)
	}
	if req.MaxRating != nil {
		conditions = append(conditions, "m.rating_average <= ?")
		args = append(args, *req.MaxRating)
	}

	baseQuery := `
SELECT
    m.id,
    m.slug,
    m.title,
    m.alt_title,
    m.author,
    m.artist,
    COALESCE(GROUP_CONCAT(DISTINCT t.name), '') AS genres,
    m.status,
    m.synopsis,
    m.cover_url,
    m.rating_average,
    m.rating_count
FROM mangas m
LEFT JOIN manga_tags mt ON m.id = mt.manga_id
LEFT JOIN tags t ON mt.tag_id = t.id
`

	// --- Build WHERE ---
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
	}

	groupBy := " GROUP BY m.id"
	baseQuery += groupBy

	// --- Sorting ---
	switch req.SortBy {
	case "rating":
		baseQuery += " ORDER BY m.rating_average DESC, m.rating_count DESC"
	case "date_updated":
		baseQuery += " ORDER BY m.updated_at DESC"
	case "relevance":
		baseQuery += " ORDER BY m.rating_count DESC, m.rating_average DESC"
	default:
		baseQuery += " ORDER BY m.rating_average DESC, m.rating_count DESC"
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

	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") as counted"

	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)

	baseQuery += " LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	// --- Execute Search Query ---
	rows, err := r.db.QueryContext(ctx, baseQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []Manga
	for rows.Next() {
		var (
			m      Manga
			alt    sql.NullString
			author sql.NullString
			artist sql.NullString
			desc   sql.NullString
			image  sql.NullString
			genres sql.NullString
			views  int64
		)
		if err := rows.Scan(
			&m.ID,
			&m.Slug,
			&m.Title,
			&alt,
			&author,
			&artist,
			&genres,
			&m.Status,
			&desc,
			&image,
			&m.RatingPoint,
			&views,
		); err != nil {
			return nil, 0, err
		}
		m.Name = m.Title
		m.Views = views
		m.Author = author.String
		m.Artist = artist.String
		m.Description = desc.String
		m.Image = image.String
		if alt.Valid && m.Slug == "" {
			m.Slug = alt.String
		}
		if genres.Valid {
			m.Genre = genres.String
		}
		results = append(results, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// --- Count Query ---
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetByID retrieves manga details by ID
func (r *Repository) GetByID(ctx context.Context, mangaID int64) (*Manga, error) {
	query := `
SELECT
    m.id,
    m.slug,
    m.title,
    m.alt_title,
    m.author,
    m.artist,
    COALESCE(GROUP_CONCAT(DISTINCT t.name), '') AS genres,
    m.status,
    m.synopsis,
    m.cover_url,
    m.rating_average,
    m.rating_count
FROM mangas m
LEFT JOIN manga_tags mt ON m.id = mt.manga_id
LEFT JOIN tags t ON mt.tag_id = t.id
WHERE m.id = ?
GROUP BY m.id
`

	var (
		m      Manga
		alt    sql.NullString
		author sql.NullString
		artist sql.NullString
		desc   sql.NullString
		image  sql.NullString
		genres sql.NullString
		views  int64
	)
	err := r.db.QueryRowContext(ctx, query, mangaID).Scan(
		&m.ID,
		&m.Slug,
		&m.Title,
		&alt,
		&author,
		&artist,
		&genres,
		&m.Status,
		&desc,
		&image,
		&m.RatingPoint,
		&views,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Name = m.Title
	m.Views = views
	m.Author = author.String
	m.Artist = artist.String
	m.Description = desc.String
	m.Image = image.String
	if alt.Valid && m.Slug == "" {
		m.Slug = alt.String
	}
	if genres.Valid {
		m.Genre = genres.String
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
SELECT
    m.id,
    m.slug,
    m.title,
    m.alt_title,
    m.author,
    m.artist,
    COALESCE(GROUP_CONCAT(DISTINCT t.name), '') AS genres,
    m.status,
    m.synopsis,
    m.cover_url,
    m.rating_average,
    m.rating_count
FROM mangas m
LEFT JOIN manga_tags mt ON m.id = mt.manga_id
LEFT JOIN tags t ON mt.tag_id = t.id
GROUP BY m.id
ORDER BY m.rating_average DESC, m.rating_count DESC, m.updated_at DESC
LIMIT ?
`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var popular []Manga
	for rows.Next() {
		var (
			m      Manga
			alt    sql.NullString
			author sql.NullString
			artist sql.NullString
			desc   sql.NullString
			image  sql.NullString
			genres sql.NullString
			views  int64
		)
		if err := rows.Scan(
			&m.ID,
			&m.Slug,
			&m.Title,
			&alt,
			&author,
			&artist,
			&genres,
			&m.Status,
			&desc,
			&image,
			&m.RatingPoint,
			&views,
		); err != nil {
			return nil, err
		}
		m.Name = m.Title
		m.Views = views
		m.Author = author.String
		m.Artist = artist.String
		m.Description = desc.String
		m.Image = image.String
		if alt.Valid && m.Slug == "" {
			m.Slug = alt.String
		}
		if genres.Valid {
			m.Genre = genres.String
		}
		popular = append(popular, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return popular, nil
}

// GetByTitle retrieves a manga by title (case-insensitive)
func (r *Repository) GetByTitle(ctx context.Context, title string) (*Manga, error) {
	query := `
SELECT id, slug, title, alt_title, author, artist, status, synopsis, cover_url, rating_average, rating_count
FROM mangas
WHERE LOWER(title) = LOWER(?)
LIMIT 1
`

	var (
		m      Manga
		alt    sql.NullString
		author sql.NullString
		artist sql.NullString
		desc   sql.NullString
		image  sql.NullString
		views  int64
	)
	err := r.db.QueryRowContext(ctx, query, title).Scan(
		&m.ID,
		&m.Slug,
		&m.Title,
		&alt,
		&author,
		&artist,
		&m.Status,
		&desc,
		&image,
		&m.RatingPoint,
		&views,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Name = m.Title
	m.Author = author.String
	m.Artist = artist.String
	m.Description = desc.String
	m.Image = image.String
	m.Views = views
	if alt.Valid && m.Slug == "" {
		m.Slug = alt.String
	}
	return &m, nil
}

// Create inserts a manga and its tags, returning the new ID.
func (r *Repository) Create(ctx context.Context, req CreateMangaRequest) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if req.Language == "" {
		req.Language = "ja"
	}

	result, err := tx.ExecContext(ctx, `
INSERT INTO mangas (slug, title, alt_title, cover_url, author, artist, status, synopsis, language, rating_average, rating_count, last_chapter)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, req.Slug, req.Title, req.AltTitle, req.CoverURL, req.Author, req.Artist, req.Status, req.Synopsis, req.Language, req.Rating, req.Views, req.LastChapter)
	if err != nil {
		return 0, err
	}

	mangaID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, genre := range req.Genres {
		genre = strings.TrimSpace(genre)
		if genre == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO NOTHING`, genre); err != nil {
			return 0, err
		}

		var tagID int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, genre).Scan(&tagID); err != nil {
			return 0, err
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO manga_tags (manga_id, tag_id) VALUES (?, ?) ON CONFLICT(manga_id, tag_id) DO NOTHING`, mangaID, tagID); err != nil {
			return 0, err
		}
	}

	if req.LastChapter > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE mangas SET last_chapter = ?, last_chapter_at = CURRENT_TIMESTAMP WHERE id = ?`, req.LastChapter, mangaID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return mangaID, nil
}
