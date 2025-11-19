package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/comment"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/history"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/library"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/chapter"
)

// WriteProcessor processes queued write operations
type WriteProcessor struct {
	queue        *WriteQueue
	mangaService *manga.Service
	db           *sql.DB
}

// NewWriteProcessor creates a new write processor
func NewWriteProcessor(queue *WriteQueue, mangaService *manga.Service, db *sql.DB) *WriteProcessor {
	return &WriteProcessor{
		queue:        queue,
		mangaService: mangaService,
		db:           db,
	}
}

// ProcessOperation processes a single write operation
func (p *WriteProcessor) ProcessOperation(ctx context.Context, op WriteOperation) error {
	switch op.Type {
	case "add_to_library":
		return p.processAddToLibrary(ctx, op)
	case "update_progress":
		return p.processUpdateProgress(ctx, op)
	case "create_review":
		return p.processCreateReview(ctx, op)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// processAddToLibrary processes an add to library operation
func (p *WriteProcessor) processAddToLibrary(ctx context.Context, op WriteOperation) error {
	status, _ := op.Data["status"].(string)
	currentChapter, _ := op.Data["current_chapter"].(int)

	// Create repository and add to library
	libraryRepo := library.NewRepository(p.db)

	// Check if already exists
	exists, err := libraryRepo.CheckLibraryExists(ctx, op.UserID, op.MangaID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already exists, skip
	}

	if currentChapter < 1 {
		currentChapter = 1
	}

	return libraryRepo.AddToLibrary(ctx, op.UserID, op.MangaID, status, currentChapter)
}

// processUpdateProgress processes an update progress operation
func (p *WriteProcessor) processUpdateProgress(ctx context.Context, op WriteOperation) error {
	currentChapter, ok := op.Data["current_chapter"].(int)
	if !ok {
		// Try float64 (JSON numbers)
		if f, ok := op.Data["current_chapter"].(float64); ok {
			currentChapter = int(f)
		} else {
			return fmt.Errorf("invalid current_chapter type")
		}
	}

	if currentChapter < 1 {
		return fmt.Errorf("invalid chapter number")
	}

	// Create repository and update progress
	libraryRepo := library.NewRepository(p.db)

	// Check if manga exists in library
	exists, err := libraryRepo.CheckLibraryExists(ctx, op.UserID, op.MangaID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("manga not in library")
	}

	chapterService := chapter.NewService(chapter.NewRepository(p.db))
	summary, err := chapterService.ValidateChapter(ctx, op.MangaID, currentChapter)
	if err != nil {
		return err
	}
	if summary == nil {
		return fmt.Errorf("chapter not found")
	}

	var chapterID *int64
	if summary.ID != 0 {
		id := summary.ID
		chapterID = &id
	}

	historyRepo := history.NewRepository(p.db)
	return historyRepo.UpdateProgress(ctx, op.UserID, op.MangaID, currentChapter, chapterID)
}

// processCreateReview processes a create review operation
func (p *WriteProcessor) processCreateReview(ctx context.Context, op WriteOperation) error {
	rating, ok := op.Data["rating"].(int)
	if !ok {
		// Try float64 (JSON numbers)
		if f, ok := op.Data["rating"].(float64); ok {
			rating = int(f)
		} else {
			return fmt.Errorf("invalid rating type")
		}
	}

	content, _ := op.Data["content"].(string)

	if rating < 1 || rating > 10 {
		return fmt.Errorf("invalid rating")
	}
	if len(content) < 10 || len(content) > 5000 {
		return fmt.Errorf("invalid content length")
	}

	// Create repositories used during review creation
	commentRepo := comment.NewRepository(p.db)
	historyRepo := history.NewRepository(p.db)

	// Check if review already exists
	existing, err := commentRepo.GetReviewByUserAndManga(ctx, op.UserID, op.MangaID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // Already exists, skip
	}

	// Check if manga is completed
	completed, err := historyRepo.IsMangaCompleted(ctx, op.UserID, op.MangaID)
	if err != nil {
		return err
	}
	if !completed {
		return fmt.Errorf("manga must be completed to write review")
	}

	_, err = commentRepo.CreateReview(ctx, op.UserID, op.MangaID, rating, content)
	return err
}

// StartProcessing starts the background processor
func (p *WriteProcessor) StartProcessing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.queue.IsEmpty() {
				// Use our processor function
				processed, failed := p.ProcessAllNow(ctx)
				if processed > 0 || failed > 0 {
					log.Printf("Processed %d operations, %d failed", processed, failed)
				}
			}
		}
	}
}

// ProcessAllNow processes all queued operations immediately
func (p *WriteProcessor) ProcessAllNow(ctx context.Context) (processed, failed int) {
	// Update process function
	originalProcessFunc := p.queue.processFunc
	p.queue.processFunc = p.ProcessOperation

	processed, failed = p.queue.ProcessAll(ctx)

	// Restore original
	p.queue.processFunc = originalProcessFunc

	return processed, failed
}
