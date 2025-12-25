package chapter

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pkgchapter "github.com/ngocan-dev/mangahub/backend/pkg/models"
)

// Repository encapsulates all database access related to chapters.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs a repository instance.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetChapters returns paginated chapters for a manga ordered by chapter number.
func (r *Repository) GetChapters(ctx context.Context, mangaID int64, limit, offset int) ([]pkgchapter.ChapterSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, manga_id, number, title, updated_at
        FROM chapters
        WHERE manga_id = ?
        ORDER BY number ASC
        LIMIT ? OFFSET ?
    `, mangaID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []pkgchapter.ChapterSummary
	for rows.Next() {
		var (
			summary   pkgchapter.ChapterSummary
			title     sql.NullString
			updatedAt sql.NullTime
		)
		if err := rows.Scan(&summary.ID, &summary.MangaID, &summary.Number, &title, &updatedAt); err != nil {
			return nil, err
		}
		summary.Title = title.String
		if updatedAt.Valid {
			t := updatedAt.Time
			summary.UpdatedAt = &t
		}
		chapters = append(chapters, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chapters, nil
}

// GetChapter returns a single chapter (including content) by number.
func (r *Repository) GetChapter(ctx context.Context, mangaID int64, chapterNumber int) (*pkgchapter.Chapter, error) {
	if chapterNumber < 1 {
		return nil, nil
	}

	var (
		chapter   pkgchapter.Chapter
		title     sql.NullString
		content   sql.NullString
		updatedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
        SELECT id, manga_id, number, title, content_url, updated_at
        FROM chapters
        WHERE manga_id = ? AND number = ?
        LIMIT 1
    `, mangaID, chapterNumber).Scan(&chapter.ID, &chapter.MangaID, &chapter.Number, &title, &content, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	chapter.Title = title.String
	if updatedAt.Valid {
		t := updatedAt.Time
		chapter.UpdatedAt = &t
	}
	chapter.Content = content.String
	return &chapter, nil
}

// GetChapterByID returns a chapter by its primary key.
func (r *Repository) GetChapterByID(ctx context.Context, chapterID int64) (*pkgchapter.Chapter, error) {
	if chapterID < 1 {
		return nil, nil
	}

	var (
		chapter   pkgchapter.Chapter
		title     sql.NullString
		content   sql.NullString
		updatedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
        SELECT id, manga_id, number, title, content_url, updated_at
        FROM chapters
        WHERE id = ?
        LIMIT 1
    `, chapterID).Scan(&chapter.ID, &chapter.MangaID, &chapter.Number, &title, &content, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	chapter.Title = title.String
	if updatedAt.Valid {
		t := updatedAt.Time
		chapter.UpdatedAt = &t
	}
	chapter.Content = content.String
	return &chapter, nil
}

// ValidateChapter checks whether a chapter exists and returns its summary when present.
func (r *Repository) ValidateChapter(ctx context.Context, mangaID int64, chapterNumber int) (*pkgchapter.ChapterSummary, error) {
	if chapterNumber < 1 {
		return nil, nil
	}

	var (
		summary   pkgchapter.ChapterSummary
		title     sql.NullString
		updatedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
        SELECT id, manga_id, number, title, updated_at
        FROM chapters
        WHERE manga_id = ? AND number = ?
        LIMIT 1
    `, mangaID, chapterNumber).Scan(&summary.ID, &summary.MangaID, &summary.Number, &title, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	summary.Title = title.String
	if updatedAt.Valid {
		t := updatedAt.Time
		summary.UpdatedAt = &t
	}
	return &summary, nil
}

// GetChapterCount returns the number of chapters available for a manga.
func (r *Repository) GetChapterCount(ctx context.Context, mangaID int64) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chapters WHERE manga_id = ?`, mangaID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CreateChapter inserts a chapter and updates manga metadata.
func (r *Repository) CreateChapter(ctx context.Context, mangaID int64, number int, title string, contentURL string, language string) (int64, error) {
	if language == "" {
		language = "ja"
	}

	now := time.Now()
	result, err := r.db.ExecContext(ctx, `
INSERT INTO chapters (manga_id, number, title, language, content_url, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(manga_id, number, language) DO UPDATE SET title=excluded.title, content_url=excluded.content_url, updated_at=excluded.updated_at
`, mangaID, number, title, language, contentURL, now)
	if err != nil {
		return 0, err
	}

	chapterID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if _, err := r.db.ExecContext(ctx, `
UPDATE mangas
SET last_chapter = ?, last_chapter_at = ?
WHERE id = ? AND (last_chapter IS NULL OR last_chapter < ?)
`, number, now, mangaID, number); err != nil {
		return 0, err
	}

	return chapterID, nil
}
