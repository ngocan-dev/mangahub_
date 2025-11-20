package history

import "time"

// UserProgress represents user's reading progress
type UserProgress struct {
	CurrentChapter   int       `json:"current_chapter"`
	CurrentChapterID *int64    `json:"current_chapter_id,omitempty"`
	LastReadAt       time.Time `json:"last_read_at"`
}

// UpdateProgressRequest represents progress update payload
type UpdateProgressRequest struct {
	CurrentChapter int `json:"current_chapter" binding:"required"`
}

// UpdateProgressResponse represents response after update
type UpdateProgressResponse struct {
	Message      string        `json:"message"`
	UserProgress *UserProgress `json:"user_progress"`
	Broadcasted  bool          `json:"broadcasted"`
}

// Activity represents user activity entry
type Activity struct {
	ActivityID    int64      `json:"activity_id"`
	UserID        int64      `json:"user_id"`
	Username      string     `json:"username"`
	ActivityType  string     `json:"activity_type"`
	MangaID       int64      `json:"manga_id"`
	MangaName     string     `json:"manga_name,omitempty"`
	MangaTitle    string     `json:"manga_title,omitempty"`
	MangaImage    string     `json:"manga_image,omitempty"`
	Rating        *int       `json:"rating,omitempty"`
	ReviewID      *int64     `json:"review_id,omitempty"`
	ReviewContent *string    `json:"review_content,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ActivityFeedResponse wraps paginated activities
type ActivityFeedResponse struct {
	Activities []Activity `json:"activities"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Pages      int        `json:"pages"`
}

// GenreStat aggregates reading per genre
type GenreStat struct {
	Genre    string `json:"genre"`
	Count    int    `json:"count"`
	Chapters int    `json:"chapters"`
}

// MonthlyStat represents monthly stats
type MonthlyStat struct {
	Year           int `json:"year"`
	Month          int `json:"month"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
}

// YearlyStat represents yearly stats
type YearlyStat struct {
	Year           int `json:"year"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
	TotalDays      int `json:"total_days"`
}

// ReadingGoal represents user reading goals
type ReadingGoal struct {
	GoalID       int64     `json:"goal_id"`
	UserID       int64     `json:"user_id"`
	GoalType     string    `json:"goal_type"`
	TargetValue  int       `json:"target_value"`
	CurrentValue int       `json:"current_value"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Completed    bool      `json:"completed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReadingStatistics aggregates user reading metrics
type ReadingStatistics struct {
	UserID                int64         `json:"user_id"`
	TotalChaptersRead     int           `json:"total_chapters_read"`
	TotalMangaRead        int           `json:"total_manga_read"`
	TotalMangaReading     int           `json:"total_manga_reading"`
	TotalMangaPlanned     int           `json:"total_manga_planned"`
	FavoriteGenres        []GenreStat   `json:"favorite_genres"`
	AverageRating         float64       `json:"average_rating"`
	TotalReadingTimeHours float64       `json:"total_reading_time_hours"`
	CurrentStreakDays     int           `json:"current_streak_days"`
	LongestStreakDays     int           `json:"longest_streak_days"`
	MonthlyStats          []MonthlyStat `json:"monthly_stats"`
	YearlyStats           []YearlyStat  `json:"yearly_stats"`
	LastCalculatedAt      time.Time     `json:"last_calculated_at"`
	Goals                 []ReadingGoal `json:"goals,omitempty"`
}

// ReadingAnalyticsRequest represents analytics filters
type ReadingAnalyticsRequest struct {
	TimePeriod   string `form:"time_period" json:"time_period"`
	Year         *int   `form:"year" json:"year"`
	Month        *int   `form:"month" json:"month"`
	IncludeGoals bool   `form:"include_goals" json:"include_goals"`
}
