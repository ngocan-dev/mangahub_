package manga

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/security"
)

var (
	ErrDatabaseError = errors.New("database error")
)

type Service struct {
	repo        *Repository
	broadcaster Broadcaster
	cache       MangaCacher       // Optional cache interface
	dbHealth    DBHealthChecker   // Database health checker
	writeQueue  WriteQueueManager // Write queue manager
}

// MangaCacher interface for manga caching
type MangaCacher interface {
	GetMangaDetail(ctx context.Context, mangaID int64) (*MangaDetail, error)
	SetMangaDetail(ctx context.Context, mangaID int64, detail *MangaDetail) error
	InvalidateMangaDetail(ctx context.Context, mangaID int64) error
	GetSearchResults(ctx context.Context, cacheKey string) (*SearchResponse, error)
	SetSearchResults(ctx context.Context, cacheKey string, response *SearchResponse) error
}

// Broadcaster interface for TCP broadcasting
type Broadcaster interface {
	BroadcastProgress(ctx context.Context, userID, novelID int64, chapter int, chapterID *int64) error
}

func NewService(db *sql.DB) *Service {
	return &Service{
		repo:        NewRepository(db),
		broadcaster: nil, // Will be set if TCP broadcaster is available
		cache:       nil, // Will be set if cache is available
	}
}

// SetCache sets the cache for the service
func (s *Service) SetCache(cache MangaCacher) {
	s.cache = cache
}

// SetDBHealth sets the database health checker
func (s *Service) SetDBHealth(checker DBHealthChecker) {
	s.dbHealth = checker
}

// SetWriteQueue sets the write queue manager
func (s *Service) SetWriteQueue(queue WriteQueueManager) {
	s.writeQueue = queue
}

// SetBroadcaster sets the TCP broadcaster for progress updates
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// Search searches for manga based on criteria
// Main Success Scenario:
// 1. User enters search query (title or author)
// 2. System queries SQLite database using LIKE patterns
// 3. System applies basic filters (genre, status) if provided
// 4. System returns paginated results with basic information
// 5. User can select manga for detailed view
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	// Validate and set defaults
	// Pagination prevents memory issues
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit to prevent memory issues and slow queries
	}
	// Prevent extremely large page numbers that could cause performance issues
	if req.Page > 10000 {
		req.Page = 10000 // Reasonable upper limit
	}

	// Step 4: Check cache first if available
	if s.cache != nil {
		cacheKey := GenerateSearchCacheKey(req)
		cached, err := s.cache.GetSearchResults(ctx, cacheKey)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Perform search
	results, total, err := s.repo.Search(ctx, req)
	if err != nil {
		// A2: Database error - System logs error and returns generic message
		// Return wrapped error that preserves original error for logging
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Calculate total pages
	pages := int(math.Ceil(float64(total) / float64(req.Limit)))
	if pages == 0 {
		pages = 1
	}

	// A1: No results found - handled by returning empty results array
	// The response will have total=0, which the handler can use to show "no results" message

	response := &SearchResponse{
		Results: results,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Pages:   pages,
	}

	// Step 2: Store in cache if available
	if s.cache != nil {
		cacheKey := GenerateSearchCacheKey(req)
		s.cache.SetSearchResults(ctx, cacheKey, response)
	}

	return response, nil
}

// GenerateSearchCacheKey generates a cache key for search request
func GenerateSearchCacheKey(req SearchRequest) string {
	// Create a unique key based on search parameters
	key := fmt.Sprintf("q:%s", req.Query)
	if len(req.Genres) > 0 {
		key += fmt.Sprintf(":genres:%v", req.Genres)
	}
	if req.Status != "" {
		key += fmt.Sprintf(":status:%s", req.Status)
	}
	if req.MinRating != nil {
		key += fmt.Sprintf(":min_rating:%.1f", *req.MinRating)
	}
	if req.MaxRating != nil {
		key += fmt.Sprintf(":max_rating:%.1f", *req.MaxRating)
	}
	if req.YearFrom != nil {
		key += fmt.Sprintf(":year_from:%d", *req.YearFrom)
	}
	if req.YearTo != nil {
		key += fmt.Sprintf(":year_to:%d", *req.YearTo)
	}
	key += fmt.Sprintf(":page:%d:limit:%d:sort:%s", req.Page, req.Limit, req.SortBy)
	return key
}

var (
	ErrMangaNotFound = errors.New("manga not found")
)

// GetDetails retrieves detailed manga information
// Main Success Scenario:
// 1. User selects manga from search results or direct URL
// 2. System retrieves manga details from database
// 3. System displays title, author, genres, description, chapter count
// 4. System shows user's current progress if logged in
// 5. User can add manga to library or update progress
// Step 1: System identifies frequently requested manga
// Step 4: Subsequent requests serve data from cache
// Resilience: Read operations return cached data when database is unavailable
func (s *Service) GetDetails(ctx context.Context, mangaID int64, userID *int64) (*MangaDetail, error) {
	// Check if database is healthy
	dbHealthy := s.dbHealth == nil || s.dbHealth.IsHealthy()

	// Step 4: Check cache first if available (only for non-user-specific data)
	// Note: We cache base manga data, but user-specific data (progress, library status) is always fresh
	if s.cache != nil && userID == nil {
		cached, err := s.cache.GetMangaDetail(ctx, mangaID)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// If database is unhealthy, try cache even for authenticated users
	if !dbHealthy && s.cache != nil {
		cached, err := s.cache.GetMangaDetail(ctx, mangaID)
		if err == nil && cached != nil {
			// Return cached data with a note that it might be stale
			return cached, nil
		}
		// If no cache available and DB is down, return error
		return nil, ErrDatabaseUnavailable
	}

	// Step 2: Retrieve manga from database
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		// If database error and we have cache, try cache as fallback
		if s.cache != nil {
			cached, cacheErr := s.cache.GetMangaDetail(ctx, mangaID)
			if cacheErr == nil && cached != nil {
				return cached, nil
			}
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Manga not found
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	// Step 3: Get chapter count
	chapterCount, err := s.repo.GetChapterCount(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	detail := &MangaDetail{
		Manga:        *manga,
		ChapterCount: chapterCount,
	}

	// Step 4: Get user progress if logged in
	if userID != nil {
		progress, err := s.repo.GetUserProgress(ctx, *userID, mangaID)
		if err != nil {
			// Log error but don't fail the request
			// In production, you might want to log this
			progress = nil
		}
		detail.UserProgress = progress

		// Get library status
		libraryStatus, err := s.repo.GetLibraryStatus(ctx, *userID, mangaID)
		if err != nil {
			// Log error but don't fail the request
			libraryStatus = nil
		}
		detail.LibraryStatus = libraryStatus
	}

	// Step 3: Calculate average rating from all reviews
	reviewStats, err := s.repo.GetReviewStats(ctx, mangaID)
	if err != nil {
		// Log error but don't fail the request
		reviewStats = nil
	}
	detail.ReviewStats = reviewStats

	// Step 2: Store in cache if available (only base data, without user-specific fields)
	if s.cache != nil && userID == nil {
		// Create a cacheable version without user-specific data
		cacheableDetail := &MangaDetail{
			Manga:        detail.Manga,
			ChapterCount: detail.ChapterCount,
			ReviewStats:  detail.ReviewStats,
			// UserProgress and LibraryStatus are nil for cached version
		}
		s.cache.SetMangaDetail(ctx, mangaID, cacheableDetail)
	}

	return detail, nil
}

var (
	ErrMangaAlreadyInLibrary = errors.New("manga already in library")
	ErrInvalidStatus         = errors.New("invalid status")
)

// Valid statuses
var validStatuses = map[string]bool{
	"plan_to_read": true,
	"reading":      true,
	"completed":    true,
	"on_hold":      true,
	"dropped":      true,
}

// AddToLibrary adds manga to user's library
// Main Success Scenario:
// 1. User clicks "Add to Library" from manga details
// 2. System presents status options (Reading, Completed, Plan to Read)
// 3. User selects initial status and current chapter
// 4. System creates user_progress record in database
// 5. System confirms addition and updates UI
func (s *Service) AddToLibrary(ctx context.Context, userID, mangaID int64, req AddToLibraryRequest) (*AddToLibraryResponse, error) {
	// Validate status
	if !validStatuses[req.Status] {
		return nil, ErrInvalidStatus
	}

	// Check if manga exists
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	// A1: Check if manga already in library
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if exists {
		return nil, ErrMangaAlreadyInLibrary
	}

	// Step 4: Add to library
	currentChapter := req.CurrentChapter
	if currentChapter < 1 {
		currentChapter = 1 // Default to chapter 1
	}

	err = s.repo.AddToLibrary(ctx, userID, mangaID, req.Status, currentChapter, req.IsFavorite)
	if err != nil {
		// A2: Database error
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Step 5: Get updated library status and progress
	libraryStatus, err := s.repo.GetLibraryStatus(ctx, userID, mangaID)
	if err != nil {
		// Log but don't fail
		libraryStatus = nil
	}

	userProgress, err := s.repo.GetUserProgress(ctx, userID, mangaID)
	if err != nil {
		// Log but don't fail
		userProgress = nil
	}

	// Step 5: Invalidate cache when data changes
	s.invalidateMangaCache(ctx, mangaID)

	return &AddToLibraryResponse{
		Message:       "manga added to library successfully",
		LibraryStatus: libraryStatus,
		UserProgress:  userProgress,
	}, nil
}

// invalidateMangaCache invalidates cache for a manga
func (s *Service) invalidateMangaCache(ctx context.Context, mangaID int64) {
	if s.cache != nil {
		s.cache.InvalidateMangaDetail(ctx, mangaID)
		// Also invalidate search results as they might include this manga
		// Note: In production, you might want more granular cache invalidation
	}
}

var (
	ErrInvalidChapterNumber = errors.New("invalid chapter number")
	ErrMangaNotInLibrary    = errors.New("manga not in library")
)

// UpdateProgress updates user's reading progress
// Main Success Scenario:
// 1. User updates current chapter number
// 2. System validates chapter number against manga metadata
// 3. System updates user_progress record with timestamp
// 4. System triggers TCP broadcast to connected clients
// 5. System confirms update to user
func (s *Service) UpdateProgress(ctx context.Context, userID, mangaID int64, req UpdateProgressRequest) (*UpdateProgressResponse, error) {
	// Step 2: Validate chapter number
	if req.CurrentChapter < 1 {
		return nil, ErrInvalidChapterNumber
	}

	// Check if manga exists
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	// Check if manga is in user's library
	exists, err := s.repo.CheckLibraryExists(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !exists {
		return nil, ErrMangaNotInLibrary
	}

	// Validate chapter number exists
	maxChapter, err := s.repo.GetMaxChapterNumber(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// A1: Invalid chapter number - System shows validation error
	if req.CurrentChapter > maxChapter {
		return nil, fmt.Errorf("%w: chapter %d exceeds maximum chapter %d", ErrInvalidChapterNumber, req.CurrentChapter, maxChapter)
	}

	// Get chapter ID if chapter exists
	valid, chapterID, err := s.repo.ValidateChapterNumber(ctx, mangaID, req.CurrentChapter)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !valid {
		// Chapter doesn't exist, but we'll still allow the update
		// (maybe user is tracking progress before chapters are uploaded)
		chapterID = nil
	}

	// Step 3: Update progress
	err = s.repo.UpdateProgress(ctx, userID, mangaID, req.CurrentChapter, chapterID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Step 4: Broadcast via TCP (A2: if unavailable, update still succeeds)
	broadcasted := false
	if s.broadcaster != nil {
		err = s.broadcaster.BroadcastProgress(ctx, userID, mangaID, req.CurrentChapter, chapterID)
		if err == nil {
			broadcasted = true
		}
		// If broadcast fails, we still return success (A2)
	}

	// Get updated progress
	userProgress, err := s.repo.GetUserProgress(ctx, userID, mangaID)
	if err != nil {
		// Log but don't fail
		userProgress = nil
	}

	// Step 5: Invalidate cache when data changes
	s.invalidateMangaCache(ctx, mangaID)

	// Step 5: Return confirmation
	return &UpdateProgressResponse{
		Message:      "progress updated successfully",
		UserProgress: userProgress,
		Broadcasted:  broadcasted,
	}, nil
}

var (
	ErrReviewAlreadyExists   = errors.New("review already exists for this manga")
	ErrMangaNotCompleted     = errors.New("manga must be in completed list to write review")
	ErrInvalidReviewRating   = errors.New("rating must be between 1 and 10")
	ErrReviewContentTooShort = errors.New("review content must be at least 10 characters")
	ErrReviewContentTooLong  = errors.New("review content must not exceed 5000 characters")
)

// CreateReview creates a new review for a manga
// Main Success Scenario:
// 1. User navigates to manga and clicks "Write Review"
// 2. User writes review text and assigns rating (1-10)
// 3. System validates review content and rating
// 4. System saves review to database with timestamp
// 5. System displays review on manga page
func (s *Service) CreateReview(ctx context.Context, userID, mangaID int64, req CreateReviewRequest) (*CreateReviewResponse, error) {
	// Step 3: Validate review content and rating
	// Invalid data formats are rejected
	if err := security.ValidateReviewRating(req.Rating); err != nil {
		return nil, ErrInvalidReviewRating
	}

	// Input length limits are enforced
	// XSS attempts are sanitized
	if err := security.ValidateReviewContent(req.Content); err != nil {
		if errors.Is(err, security.ErrInputTooShort) {
			return nil, ErrReviewContentTooShort
		}
		if errors.Is(err, security.ErrInputTooLong) {
			return nil, ErrReviewContentTooLong
		}
		if errors.Is(err, security.ErrContainsSQLInjection) {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Sanitize review content to prevent XSS
	// XSS attempts are sanitized
	sanitizedContent := security.SanitizeReviewContent(req.Content)

	// Check if manga exists
	manga, err := s.repo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if manga == nil {
		return nil, ErrMangaNotFound
	}

	// Precondition: User has manga in completed list
	completed, err := s.repo.CheckMangaInCompletedLibrary(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if !completed {
		return nil, ErrMangaNotCompleted
	}

	// Check if user already has a review for this manga
	existingReview, err := s.repo.GetReviewByUserAndManga(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if existingReview != nil {
		return nil, ErrReviewAlreadyExists
	}

	// Step 4: Save review to database with timestamp
	// Use sanitized content to prevent XSS
	_, err = s.repo.CreateReview(ctx, userID, mangaID, req.Rating, sanitizedContent)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get the created review with user info
	review, err := s.repo.GetReviewByUserAndManga(ctx, userID, mangaID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if review == nil {
		return nil, fmt.Errorf("%w: review not found after creation", ErrDatabaseError)
	}

	// Step 5: Invalidate cache when data changes (review stats changed)
	s.invalidateMangaCache(ctx, mangaID)

	return &CreateReviewResponse{
		Message: "review created successfully",
		Review:  review,
	}, nil
}

// GetReviews retrieves reviews for a manga with pagination
// Main Success Scenario:
// 1. User views manga details page
// 2. System retrieves all reviews for the manga
// 3. System calculates average rating from all reviews
// 4. System displays reviews sorted by helpfulness or date
// 5. User can read individual reviews and ratings
func (s *Service) GetReviews(ctx context.Context, mangaID int64, page, limit int, sortBy string) (*GetReviewsResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Step 2: Retrieve reviews for the manga
	// Step 4: Sort by helpfulness or date
	reviews, total, err := s.repo.GetReviews(ctx, mangaID, page, limit, sortBy)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Step 3: Calculate average rating from all reviews
	reviewStats, err := s.repo.GetReviewStats(ctx, mangaID)
	if err != nil {
		// Log error but don't fail the request
		reviewStats = nil
	}

	// Calculate pages
	pages := (total + limit - 1) / limit
	if pages < 1 {
		pages = 1
	}

	response := &GetReviewsResponse{
		Reviews: reviews,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Pages:   pages,
	}

	// Include review stats in response
	if reviewStats != nil {
		response.AverageRating = reviewStats.AverageRating
		response.TotalReviews = reviewStats.TotalReviews
	}

	return response, nil
}

// GetFriendsActivityFeed retrieves activity feed from friends
// Main Success Scenario:
// 1. User accesses friends activity page
// 2. System retrieves recent activities from friends
// 3. System displays activities (completed manga, reviews, ratings)
// 4. Activities are sorted by recency
// 5. User can click through to view details
func (s *Service) GetFriendsActivityFeed(ctx context.Context, userID int64, page, limit int) (*ActivityFeedResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Step 2: Retrieve recent activities from friends
	// Step 4: Activities are sorted by recency (handled in repository)
	activities, total, err := s.repo.GetFriendsActivities(ctx, userID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Calculate pages
	pages := (total + limit - 1) / limit
	if pages < 1 {
		pages = 1
	}

	return &ActivityFeedResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		Limit:      limit,
		Pages:      pages,
	}, nil
}

// GetReadingStatistics retrieves reading statistics for a user
// Main Success Scenario:
// 1. System analyzes user's reading progress data
// 2. System calculates total chapters read, favorite genres
// 3. System determines reading patterns and trends
// 4. System generates monthly/yearly statistics
// 5. Statistics are cached for performance
func (s *Service) GetReadingStatistics(ctx context.Context, userID int64, forceRecalculate bool) (*ReadingStatistics, error) {
	// Step 5: Check cache first (unless force recalculate)
	if !forceRecalculate {
		cached, err := s.repo.GetCachedReadingStatistics(ctx, userID)
		if err == nil && cached != nil {
			// Check if cache is still valid (less than 1 hour old)
			age := time.Since(cached.LastCalculatedAt)
			if age < time.Hour {
				return cached, nil
			}
		}
	}

	// Step 1-4: Calculate statistics
	stats, err := s.repo.CalculateReadingStatistics(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Step 5: Cache the statistics
	if err := s.repo.SaveReadingStatistics(ctx, stats); err != nil {
		// Log error but don't fail the request
		// Statistics are still returned even if caching fails
	}

	return stats, nil
}

// GetReadingAnalytics retrieves reading analytics with optional filters
// Main Success Scenario:
// 1. User accesses statistics page
// 2. System retrieves cached statistics or generates new ones
// 3. System displays charts and graphs of reading activity
// 4. User can view different time periods and breakdowns
// 5. System shows reading goals progress if set
func (s *Service) GetReadingAnalytics(ctx context.Context, userID int64, req ReadingAnalyticsRequest) (*ReadingStatistics, error) {
	// Step 2: Retrieve cached statistics or generate new ones
	forceRecalculate := false
	stats, err := s.GetReadingStatistics(ctx, userID, forceRecalculate)
	if err != nil {
		return nil, err
	}

	// Step 4: Filter by time period if specified
	if req.TimePeriod != "" && req.TimePeriod != "all_time" {
		// Filter monthly and yearly stats based on period
		var filteredMonthly []MonthlyStat
		var filteredYearly []YearlyStat

		switch req.TimePeriod {
		case "year":
			// Show last 12 months
			filteredMonthly = stats.MonthlyStats
			if len(filteredMonthly) > 12 {
				filteredMonthly = filteredMonthly[:12]
			}
			filteredYearly = stats.YearlyStats
			if len(filteredYearly) > 1 {
				filteredYearly = filteredYearly[:1]
			}
		case "month":
			// Show last 30 days (approximately 1 month)
			if req.Year != nil && req.Month != nil {
				// Specific month
				for _, m := range stats.MonthlyStats {
					if m.Year == *req.Year && m.Month == *req.Month {
						filteredMonthly = []MonthlyStat{m}
						break
					}
				}
			} else {
				// Current month
				filteredMonthly = stats.MonthlyStats
				if len(filteredMonthly) > 1 {
					filteredMonthly = filteredMonthly[:1]
				}
			}
		case "week":
			// Show last 7 days (approximate from monthly data)
			filteredMonthly = stats.MonthlyStats
			if len(filteredMonthly) > 1 {
				filteredMonthly = filteredMonthly[:1]
			}
		default:
			filteredMonthly = stats.MonthlyStats
			filteredYearly = stats.YearlyStats
		}

		stats.MonthlyStats = filteredMonthly
		stats.YearlyStats = filteredYearly
	}

	// Step 5: Include reading goals if requested
	if req.IncludeGoals {
		// Update goal progress first
		if err := s.repo.UpdateReadingGoalProgress(ctx, userID); err != nil {
			// Log error but don't fail
		}

		// Get active goals
		goals, err := s.repo.GetActiveReadingGoals(ctx, userID)
		if err == nil {
			stats.Goals = goals
		}
	}

	return stats, nil
}
