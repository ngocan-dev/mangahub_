package api

import (
	"context"
	"net/url"
	"strconv"
)

// GenreStat aggregates reading per genre.
type GenreStat struct {
	Genre    string `json:"genre"`
	Count    int    `json:"count"`
	Chapters int    `json:"chapters"`
}

// MonthlyStat captures monthly reading metrics.
type MonthlyStat struct {
	Year           int `json:"year"`
	Month          int `json:"month"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
}

// YearlyStat captures yearly reading metrics.
type YearlyStat struct {
	Year           int `json:"year"`
	ChaptersRead   int `json:"chapters_read"`
	MangaCompleted int `json:"manga_completed"`
	MangaStarted   int `json:"manga_started"`
	TotalDays      int `json:"total_days"`
}

// ReadingGoal tracks progress toward a configured goal.
type ReadingGoal struct {
	GoalID       int64  `json:"goal_id"`
	GoalType     string `json:"goal_type"`
	TargetValue  int    `json:"target_value"`
	CurrentValue int    `json:"current_value"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Completed    bool   `json:"completed"`
}

// ReadingStatistics mirrors the backend response for reading analytics.
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
	LastCalculatedAt      string        `json:"last_calculated_at"`
	Goals                 []ReadingGoal `json:"goals,omitempty"`
}

// ReadingAnalyticsRequest captures filters for analytics.
type ReadingAnalyticsRequest struct {
	TimePeriod   string
	Year         *int
	Month        *int
	IncludeGoals bool
}

// GetReadingStatistics retrieves cached or freshly calculated statistics.
func (c *Client) GetReadingStatistics(ctx context.Context, force bool) (*ReadingStatistics, error) {
	path := "/statistics/reading"
	if force {
		path += "?force=true"
	}

	var stats ReadingStatistics
	if err := c.doRequest(ctx, "GET", path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetReadingAnalytics retrieves analytics with the provided filters.
func (c *Client) GetReadingAnalytics(ctx context.Context, req ReadingAnalyticsRequest) (*ReadingStatistics, error) {
	values := url.Values{}
	if req.TimePeriod != "" {
		values.Set("time_period", req.TimePeriod)
	}
	if req.Year != nil {
		values.Set("year", strconv.Itoa(*req.Year))
	}
	if req.Month != nil {
		values.Set("month", strconv.Itoa(*req.Month))
	}
	if req.IncludeGoals {
		values.Set("include_goals", "true")
	}

	path := "/analytics/reading"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var stats ReadingStatistics
	if err := c.doRequest(ctx, "GET", path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
