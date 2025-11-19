package history

import "context"

// CompletionChecker exposes completion verification needed by other domains.
type CompletionChecker interface {
	HasCompletedManga(ctx context.Context, userID, mangaID int64) (bool, error)
}
