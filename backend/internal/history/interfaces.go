package history

import (
	"context"

	domainhistory "github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
)

// UserProgress exposes the canonical history progress model to other domains.
type UserProgress = domainhistory.UserProgress

// GenreStat exposes aggregated reading data per genre.
type GenreStat = domainhistory.GenreStat

// MonthlyStat exposes aggregated monthly reading data.
type MonthlyStat = domainhistory.MonthlyStat

// YearlyStat exposes aggregated yearly reading data.
type YearlyStat = domainhistory.YearlyStat

// ReadingGoal exposes goal tracking data.
type ReadingGoal = domainhistory.ReadingGoal

// ReadingStatistics exposes aggregated statistics for a user.
type ReadingStatistics = domainhistory.ReadingStatistics

// ReadingAnalyticsRequest exposes analytics filters for consumers.
type ReadingAnalyticsRequest = domainhistory.ReadingAnalyticsRequest

// Repository defines the persistence operations required by the history domain.
type Repository interface {
	GetUserProgress(ctx context.Context, userID, mangaID int64) (*domainhistory.UserProgress, error)
	UpdateProgress(ctx context.Context, userID, mangaID int64, chapter int, chapterID *int64) error
	IsMangaCompleted(ctx context.Context, userID, mangaID int64) (bool, error)
	CalculateReadingStatistics(ctx context.Context, userID int64) (*domainhistory.ReadingStatistics, error)
	SaveReadingStatistics(ctx context.Context, stats *domainhistory.ReadingStatistics) error
	GetCachedReadingStatistics(ctx context.Context, userID int64) (*domainhistory.ReadingStatistics, error)
	UpdateReadingGoalProgress(ctx context.Context, userID int64) error
	GetActiveReadingGoals(ctx context.Context, userID int64) ([]domainhistory.ReadingGoal, error)
}

// CompletionChecker exposes completion verification needed by other domains.
type CompletionChecker interface {
	HasCompletedManga(ctx context.Context, userID, mangaID int64) (bool, error)
}
