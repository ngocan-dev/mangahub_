package chapter

import "context"

// Service exposes chapter queries
type Service struct {
	repo *Repository
}

// NewService creates service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetChapterCount returns total chapters
func (s *Service) GetChapterCount(ctx context.Context, mangaID int64) (int, error) {
	return s.repo.GetChapterCount(ctx, mangaID)
}

// GetMaxChapterNumber returns highest chapter
func (s *Service) GetMaxChapterNumber(ctx context.Context, mangaID int64) (int, error) {
	return s.repo.GetMaxChapterNumber(ctx, mangaID)
}

// ValidateChapterNumber validates chapter presence
func (s *Service) ValidateChapterNumber(ctx context.Context, mangaID int64, chapter int) (bool, *int64, error) {
	return s.repo.ValidateChapterNumber(ctx, mangaID, chapter)
}
