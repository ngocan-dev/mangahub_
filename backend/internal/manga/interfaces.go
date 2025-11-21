package manga

import (
	"context"

	domain "github.com/ngocan-dev/mangahub/backend/domain/manga"
)

// Search exposes search capability
type Search interface {
	Search(ctx context.Context, req domain.SearchRequest) (*domain.SearchResponse, error)
}

// GetDetails exposes manga detail capability
type GetDetails interface {
	GetDetails(ctx context.Context, mangaID int64, userID *int64) (*domain.MangaDetail, error)
}

// GetByID exposes basic retrieval
type GetByID interface {
	GetByID(ctx context.Context, mangaID int64) (*domain.Manga, error)
}
