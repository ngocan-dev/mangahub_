package chapter

import (
	"context"

	repository "github.com/ngocan-dev/mangahub/manga-backend/internal/repository/chapter"
	pkgchapter "github.com/ngocan-dev/mangahub/manga-backend/pkg/models/chapter"
)

// Service exposes higher-level chapter use cases.
type Service struct {
	repo *repository.Repository
}

// NewService constructs a chapter service.
func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// GetChapters returns a paginated slice of chapter summaries.
func (s *Service) GetChapters(ctx context.Context, mangaID int64, limit, offset int) ([]pkgchapter.ChapterSummary, error) {
	return s.repo.GetChapters(ctx, mangaID, limit, offset)
}

// GetChapter returns a single chapter with its content payload.
func (s *Service) GetChapter(ctx context.Context, mangaID int64, chapterNumber int) (*pkgchapter.Chapter, error) {
	return s.repo.GetChapter(ctx, mangaID, chapterNumber)
}

// ValidateChapter ensures a chapter exists and returns its summary when found.
func (s *Service) ValidateChapter(ctx context.Context, mangaID int64, chapterNumber int) (*pkgchapter.ChapterSummary, error) {
	return s.repo.ValidateChapter(ctx, mangaID, chapterNumber)
}

// GetChapterCount returns the total number of chapters for a manga.
func (s *Service) GetChapterCount(ctx context.Context, mangaID int64) (int, error) {
	return s.repo.GetChapterCount(ctx, mangaID)
}
