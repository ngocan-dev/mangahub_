package manga

import "time"

// Manga represents a manga/novel entity
type Manga struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Genre       string    `json:"genre"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	RatingPoint float64   `json:"rating_point"`
	DateUpdated time.Time `json:"date_updated"`
}

// SearchRequest represents search criteria
type SearchRequest struct {
	Query     string   `form:"q" json:"query"`               // Search query (title, author, or description)
	Genres    []string `form:"genres" json:"genres"`         // Optional multiple genre filters
	Status    string   `form:"status" json:"status"`         // Optional status filter
	MinRating *float64 `form:"min_rating" json:"min_rating"` // Optional minimum rating (0-5)
	MaxRating *float64 `form:"max_rating" json:"max_rating"` // Optional maximum rating (0-5)
	YearFrom  *int     `form:"year_from" json:"year_from"`   // Optional year filter (from)
	YearTo    *int     `form:"year_to" json:"year_to"`       // Optional year filter (to)
	Page      int      `form:"page" json:"page"`             // Page number (default: 1)
	Limit     int      `form:"limit" json:"limit"`           // Results per page (default: 20)
	SortBy    string   `form:"sort_by" json:"sort_by"`       // Sort by: relevance, rating, date_updated (default: relevance)
}

// SearchResponse represents paginated search results
type SearchResponse struct {
	Results []Manga `json:"results"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Pages   int     `json:"pages"` // Total number of pages
}

// MangaDetail represents detailed manga information
type MangaDetail struct {
	Manga
	ChapterCount  int            `json:"chapter_count"`
	UserProgress  *UserProgress  `json:"user_progress,omitempty"`  // Only if user is logged in
	LibraryStatus *LibraryStatus `json:"library_status,omitempty"` // Only if user is logged in
	ReviewStats   *ReviewStats   `json:"review_stats,omitempty"`   // Review statistics
}

// ReviewStats represents review statistics for a manga
type ReviewStats struct {
	AverageRating float64 `json:"average_rating"` // Average rating from all reviews (1-10)
	TotalReviews  int     `json:"total_reviews"`  // Total number of reviews
}

// UserProgress represents user's reading progress
type UserProgress struct {
	CurrentChapter   int       `json:"current_chapter"`
	CurrentChapterID *int64    `json:"current_chapter_id,omitempty"`
	LastReadAt       time.Time `json:"last_read_at"`
}

// LibraryStatus represents user's library status for the manga
type LibraryStatus struct {
	Status      string     `json:"status"` // plan_to_read, reading, completed, on_hold, dropped
	IsFavorite  bool       `json:"is_favorite"`
	Rating      *int       `json:"rating,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AddToLibraryRequest represents request to add manga to library
type AddToLibraryRequest struct {
	Status         string `json:"status" binding:"required"` // plan_to_read, reading, completed, on_hold, dropped
	CurrentChapter int    `json:"current_chapter"`           // Optional: current chapter number (default: 0 or 1)
	IsFavorite     bool   `json:"is_favorite"`               // Optional: mark as favorite
}

// AddToLibraryResponse represents response after adding to library
type AddToLibraryResponse struct {
	Message       string         `json:"message"`
	LibraryStatus *LibraryStatus `json:"library_status"`
	UserProgress  *UserProgress  `json:"user_progress,omitempty"`
}

// UpdateProgressRequest represents request to update reading progress
type UpdateProgressRequest struct {
	CurrentChapter int `json:"current_chapter" binding:"required"` // Chapter number to update to
}

// UpdateProgressResponse represents response after updating progress
type UpdateProgressResponse struct {
	Message      string        `json:"message"`
	UserProgress *UserProgress `json:"user_progress"`
	Broadcasted  bool          `json:"broadcasted"` // Whether progress was broadcasted via TCP
}

// Review represents a manga review
type Review struct {
	ReviewID  int64     `json:"review_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	MangaID   int64     `json:"manga_id"`
	Rating    int       `json:"rating"`  // Rating from 1-10
	Content   string    `json:"content"` // Review text
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// CreateReviewRequest represents request to create a review
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=10"`     // Rating from 1-10
	Content string `json:"content" binding:"required,min=10,max=5000"` // Review text (10-5000 characters)
}

// CreateReviewResponse represents response after creating a review
type CreateReviewResponse struct {
	Message string  `json:"message"`
	Review  *Review `json:"review"`
}

// GetReviewsResponse represents paginated reviews
type GetReviewsResponse struct {
	Reviews       []Review `json:"reviews"`
	Total         int      `json:"total"`
	Page          int      `json:"page"`
	Limit         int      `json:"limit"`
	Pages         int      `json:"pages"`
	AverageRating float64  `json:"average_rating"` // Average rating from all reviews
	TotalReviews  int      `json:"total_reviews"`  // Total number of reviews
}

// Activity represents a user activity in the friends feed
type Activity struct {
	ActivityID    int64      `json:"activity_id"`
	UserID        int64      `json:"user_id"`
	Username      string     `json:"username"`
	ActivityType  string     `json:"activity_type"` // completed_manga, review, rating
	MangaID       int64      `json:"manga_id"`
	MangaName     string     `json:"manga_name,omitempty"`
	MangaTitle    string     `json:"manga_title,omitempty"`
	MangaImage    string     `json:"manga_image,omitempty"`
	Rating        *int       `json:"rating,omitempty"`         // For rating and review activities
	ReviewID      *int64     `json:"review_id,omitempty"`      // For review activities
	ReviewContent *string    `json:"review_content,omitempty"` // For review activities (truncated)
	CompletedAt   *time.Time `json:"completed_at,omitempty"`   // For completed manga
	CreatedAt     time.Time  `json:"created_at"`               // When the activity occurred
}

// ActivityFeedResponse represents paginated activity feed
type ActivityFeedResponse struct {
	Activities []Activity `json:"activities"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Pages      int        `json:"pages"`
}

// GenreStat represents genre statistics
type GenreStat struct {
	Genre    string `json:"genre"`
	Count    int    `json:"count"`    // Number of manga in this genre
	Chapters int    `json:"chapters"` // Total chapters read in this genre
}

// MonthlyStat represents monthly reading statistics
type MonthlyStat struct {
	Year           int `json:"year"`
	Month          int `json:"month"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
}

// YearlyStat represents yearly reading statistics
type YearlyStat struct {
	Year           int `json:"year"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
	TotalDays      int `json:"total_days"` // Days with reading activity
}

// ReadingStatistics represents user's reading statistics
type ReadingStatistics struct {
	UserID                int64         `json:"user_id"`
	TotalChaptersRead     int           `json:"total_chapters_read"`
	TotalMangaRead        int           `json:"total_manga_read"`         // Completed manga
	TotalMangaReading     int           `json:"total_manga_reading"`      // Currently reading
	TotalMangaPlanned     int           `json:"total_manga_planned"`      // Plan to read
	FavoriteGenres        []GenreStat   `json:"favorite_genres"`          // Top genres by count
	AverageRating         float64       `json:"average_rating"`           // Average rating given
	TotalReadingTimeHours float64       `json:"total_reading_time_hours"` // Estimated
	CurrentStreakDays     int           `json:"current_streak_days"`      // Current reading streak
	LongestStreakDays     int           `json:"longest_streak_days"`      // Longest reading streak
	MonthlyStats          []MonthlyStat `json:"monthly_stats"`            // Last 12 months
	YearlyStats           []YearlyStat  `json:"yearly_stats"`             // Last 5 years
	LastCalculatedAt      time.Time     `json:"last_calculated_at"`
	Goals                 []ReadingGoal `json:"goals,omitempty"` // Active reading goals
}

// ReadingGoal represents a user's reading goal
type ReadingGoal struct {
	GoalID       int64     `json:"goal_id"`
	UserID       int64     `json:"user_id"`
	GoalType     string    `json:"goal_type"`     // chapters, manga, reading_time
	TargetValue  int       `json:"target_value"`  // Target number
	CurrentValue int       `json:"current_value"` // Current progress
	PeriodType   string    `json:"period_type"`   // daily, weekly, monthly, yearly
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Status       string    `json:"status"`   // active, completed, failed
	Progress     float64   `json:"progress"` // Percentage (0-100)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReadingAnalyticsRequest represents request for analytics with filters
type ReadingAnalyticsRequest struct {
	TimePeriod   string `form:"period" json:"period"`               // all_time, year, month, week
	Year         *int   `form:"year" json:"year"`                   // Specific year (optional)
	Month        *int   `form:"month" json:"month"`                 // Specific month (optional)
	IncludeGoals bool   `form:"include_goals" json:"include_goals"` // Include reading goals
}
