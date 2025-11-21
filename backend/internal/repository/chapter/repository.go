package chapter

import (
	"context"
	"database/sql"
	"errors"

	pkgchapter "github.com/ngocan-dev/mangahub/manga-backend/pkg/models"
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
        SELECT Chapter_Id, Novel_Id, Chapter_Number, Chapter_Title, Date_Updated
        FROM Chapters
        WHERE Novel_Id = ?
        ORDER BY Chapter_Number ASC
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
        SELECT Chapter_Id, Novel_Id, Chapter_Number, Chapter_Title, Content, Date_Updated
        FROM Chapters
        WHERE Novel_Id = ? AND Chapter_Number = ?
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
        SELECT Chapter_Id, Novel_Id, Chapter_Number, Chapter_Title, Date_Updated
        FROM Chapters
        WHERE Novel_Id = ? AND Chapter_Number = ?
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
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM Chapters WHERE Novel_Id = ?`, mangaID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
