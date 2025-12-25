package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"
)

// Repository handles reading history persistence
type Repository struct {
	db *sql.DB
}

// NewRepository builds history repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) tableExists(ctx context.Context, name string) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetReadingSummary returns lean statistics for a user.
func (r *Repository) GetReadingSummary(ctx context.Context, userID int64) (*ReadingSummary, error) {
	query := `
WITH library_stats AS (
    SELECT
        COUNT(DISTINCT manga_id) AS total_manga
    FROM libraries
    WHERE user_id = ?
),
history_stats AS (
    SELECT
        MAX(created_at) AS last_read_at,
        SUM(CASE WHEN event_type = 'finished_chapter' THEN 1 ELSE 0 END) AS chapters_read
    FROM reading_history
    WHERE user_id = ?
),
progress_stats AS (
    SELECT MAX(last_read_at) AS last_read_at
    FROM reading_progress
    WHERE user_id = ?
)
SELECT
    COALESCE(ls.total_manga, 0) AS total_manga,
    COALESCE(hs.chapters_read, 0) AS total_chapters_read,
    0 AS reading_streak,
    COALESCE(hs.last_read_at, ps.last_read_at)
FROM library_stats ls
CROSS JOIN history_stats hs
CROSS JOIN progress_stats ps
LIMIT 1
`
	var summary ReadingSummary
	var lastReadAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, userID, userID, userID).Scan(
		&summary.TotalManga,
		&summary.TotalChaptersRead,
		&summary.ReadingStreak,
		&lastReadAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return &summary, nil
		}
		log.Printf("history.repository.GetReadingSummary: query error user_id=%d err=%v", userID, err)
		return nil, err
	}
	if lastReadAt.Valid {
		summary.LastReadAt = &lastReadAt.Time
	}
	return &summary, nil
}

// GetReadingAnalyticsBuckets aggregates daily/weekly/monthly analytics.
func (r *Repository) GetReadingAnalyticsBuckets(ctx context.Context, userID int64) (*ReadingAnalyticsResponse, error) {
	resp := &ReadingAnalyticsResponse{
		Daily:   []ReadingAnalyticsPoint{},
		Weekly:  []ReadingAnalyticsPoint{},
		Monthly: []ReadingAnalyticsPoint{},
	}

	type bucketQuery struct {
		dest *[]ReadingAnalyticsPoint
		sql  string
	}

	queries := []bucketQuery{
		{
			dest: &resp.Daily,
			sql: `
            SELECT date(created_at) AS bucket_date, COUNT(*) AS chapters_read
            FROM reading_history
            WHERE user_id = ?
            GROUP BY bucket_date
            ORDER BY bucket_date DESC
            LIMIT 30
            `,
		},
		{
			dest: &resp.Weekly,
			sql: `
            SELECT date(created_at, 'weekday 0', '-6 days') AS bucket_date, COUNT(*) AS chapters_read
            FROM reading_history
            WHERE user_id = ?
            GROUP BY bucket_date
            ORDER BY bucket_date DESC
            LIMIT 12
            `,
		},
		{
			dest: &resp.Monthly,
			sql: `
            SELECT strftime('%Y-%m-01', created_at) AS bucket_date, COUNT(*) AS chapters_read
            FROM reading_history
            WHERE user_id = ?
            GROUP BY bucket_date
            ORDER BY bucket_date DESC
            LIMIT 12
            `,
		},
	}

	for _, q := range queries {
		rows, err := r.db.QueryContext(ctx, q.sql, userID)
		if err != nil {
			log.Printf("history.repository.GetReadingAnalyticsBuckets: query error user_id=%d err=%v", userID, err)
			return nil, err
		}
		for rows.Next() {
			var bucket sql.NullString
			var chapters int
			if err := rows.Scan(&bucket, &chapters); err != nil {
				log.Printf("history.repository.GetReadingAnalyticsBuckets: scan error user_id=%d err=%v", userID, err)
				rows.Close()
				return nil, err
			}
			if !bucket.Valid {
				continue
			}
			*q.dest = append(*q.dest, ReadingAnalyticsPoint{
				Date:         bucket.String,
				ChaptersRead: chapters,
			})
		}
		if err := rows.Err(); err != nil {
			log.Printf("history.repository.GetReadingAnalyticsBuckets: rows error user_id=%d err=%v", userID, err)
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	return resp, nil
}

// GetUserProgress retrieves progress record
func (r *Repository) GetUserProgress(ctx context.Context, userID, mangaID int64) (*UserProgress, error) {
	query := `
        SELECT
            COALESCE(c.number, 0) as current_chapter,
            rp.current_chapter_id,
            COALESCE(rp.progress_percent, 0) as progress_percent,
            rp.last_read_at
        FROM reading_progress rp
        LEFT JOIN chapters c ON rp.current_chapter_id = c.id
        WHERE rp.user_id = ? AND rp.manga_id = ?
    `
	var progress UserProgress
	var chapterID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&progress.CurrentChapter,
		&chapterID,
		&progress.ProgressPercent,
		&progress.LastReadAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if chapterID.Valid {
		progress.CurrentChapterID = &chapterID.Int64
	}
	return &progress, nil
}

// UpdateProgress updates progress table
func (r *Repository) UpdateProgress(ctx context.Context, userID, mangaID int64, chapter int, chapterID *int64, progressPercent float64) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE reading_progress
SET current_chapter_id = ?, progress_percent = ?, last_read_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND manga_id = ?
`, chapterID, progressPercent, userID, mangaID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows > 0 {
		return nil
	}
	_, err = r.db.ExecContext(ctx, `
INSERT INTO reading_progress (user_id, manga_id, current_chapter_id, last_read_at, progress_percent, current_page)
VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?, 0)
ON CONFLICT(user_id, manga_id) DO UPDATE SET
    current_chapter_id = excluded.current_chapter_id,
    progress_percent = excluded.progress_percent,
    last_read_at = excluded.last_read_at
`, userID, mangaID, chapterID, progressPercent)
	return err
}

// IsMangaCompleted checks whether the user completed the manga in their library
func (r *Repository) IsMangaCompleted(ctx context.Context, userID, mangaID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM libraries
WHERE user_id = ? AND manga_id = ? AND status = 'completed'
`, userID, mangaID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFriends retrieves accepted friend IDs
func (r *Repository) GetFriends(ctx context.Context, userID int64) ([]int64, error) {
	exists, err := r.tableExists(ctx, "friends")
	if err != nil {
		return nil, err
	}
	if !exists {
		log.Printf("history.repository.GetFriends: friends table missing, returning empty list for user_id=%d", userID)
		return []int64{}, nil
	}
	query := `
        SELECT friend_id FROM friends WHERE user_id = ? AND status = 'accepted'
        UNION
        SELECT user_id FROM friends WHERE friend_id = ? AND status = 'accepted'
    `
	log.Printf("history.repository.GetFriends: querying friends for user_id=%d", userID)
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// RecordActivity stores an activity row for the activity feed.
func (r *Repository) RecordActivity(ctx context.Context, userID int64, activityType string, mangaID *int64, payload map[string]interface{}) error {
	exists, err := r.tableExists(ctx, "activities")
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("history.repository.RecordActivity: activities table missing user_id=%d type=%s", userID, activityType)
		return nil
	}
	payloadJSON := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			payloadJSON = string(b)
		}
	}
	_, err = r.db.ExecContext(ctx, `
        INSERT INTO activities (user_id, type, manga_id, payload)
        VALUES (?, ?, ?, ?)
    `, userID, activityType, mangaID, payloadJSON)
	return err
}

// GetFriendsActivities returns friend feed entries
func (r *Repository) GetFriendsActivities(ctx context.Context, userID int64, page, limit int) ([]Activity, int, error) {
	log.Printf("history.repository.GetFriendsActivities: start user_id=%d page=%d limit=%d", userID, page, limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	countQuery := `
        SELECT COUNT(*)
        FROM activities a
        JOIN friends f ON f.friend_id = a.user_id
        WHERE f.user_id = ?
    `
	log.Printf("history.repository.GetFriendsActivities: count_sql=%s", countQuery)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		log.Printf("history.repository.GetFriendsActivities: count query user_id=%d err=%v", userID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return []Activity{}, 0, nil
		}
		return nil, 0, err
	}
	if total == 0 {
		return []Activity{}, 0, nil
	}

	query := `
        SELECT
            a.id,
            a.user_id,
            u.username,
            a.type,
            COALESCE(a.manga_id, 0),
            m.title as manga_title,
            m.cover_url as manga_image,
            a.payload,
            a.created_at
        FROM activities a
        JOIN friends f ON f.friend_id = a.user_id
        JOIN users u ON u.id = a.user_id
        LEFT JOIN mangas m ON m.id = a.manga_id
        WHERE f.user_id = ?
        ORDER BY a.created_at DESC
        LIMIT ? OFFSET ?
    `
	log.Printf("history.repository.GetFriendsActivities: feed_sql=%s", query)
	log.Printf("history.repository.GetFriendsActivities: query user_id=%d limit=%d offset=%d", userID, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []Activity{}, total, nil
		}
		log.Printf("history.repository.GetFriendsActivities: list query failed user_id=%d err=%v", userID, err)
		return nil, 0, err
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var (
			activity   Activity
			payloadRaw sql.NullString
		)
		if err := rows.Scan(
			&activity.ActivityID,
			&activity.UserID,
			&activity.Username,
			&activity.ActivityType,
			&activity.MangaID,
			&activity.MangaTitle,
			&activity.MangaImage,
			&payloadRaw,
			&activity.CreatedAt,
		); err != nil {
			log.Printf("history.repository.GetFriendsActivities: scan error user_id=%d err=%v", userID, err)
			continue
		}
		if payloadRaw.Valid && payloadRaw.String != "" {
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(payloadRaw.String), &payload); err == nil {
				activity.Payload = payload
			} else {
				log.Printf("history.repository.GetFriendsActivities: payload parse error user_id=%d payload=%s err=%v", userID, payloadRaw.String, err)
			}
		}
		activities = append(activities, activity)
	}

	return activities, total, rows.Err()
}

// CalculateReadingStatistics aggregates stats
func (r *Repository) CalculateReadingStatistics(ctx context.Context, userID int64) (*ReadingStatistics, error) {
	stats := &ReadingStatistics{UserID: userID}

	log.Printf("history.repository.CalculateReadingStatistics: aggregating stats for user_id=%d", userID)
	err := r.db.QueryRowContext(ctx, `
        SELECT
            (SELECT COALESCE(COUNT(*), 0) FROM reading_history WHERE user_id = ? AND event_type = 'finished_chapter') AS total_chapters_read,
            COUNT(DISTINCT CASE WHEN lib.status = 'completed' THEN lib.manga_id END) as total_manga_read,
            COUNT(DISTINCT CASE WHEN lib.status = 'reading' THEN lib.manga_id END) as total_manga_reading,
            COUNT(DISTINCT CASE WHEN lib.status = 'plan_to_read' THEN lib.manga_id END) as total_manga_planned,
            COALESCE(AVG(rt.score), 0) as average_rating
        FROM libraries lib
        LEFT JOIN ratings rt ON lib.user_id = rt.user_id AND lib.manga_id = rt.manga_id
        WHERE lib.user_id = ?
    `, userID, userID).Scan(
		&stats.TotalChaptersRead,
		&stats.TotalMangaRead,
		&stats.TotalMangaReading,
		&stats.TotalMangaPlanned,
		&stats.AverageRating,
	)
	if err != nil {
		return nil, err
	}

	stats.AverageRating = float64(int(stats.AverageRating*100+0.5)) / 100

	stats.MonthlyStats = []MonthlyStat{}
	log.Printf("history.repository.CalculateReadingStatistics: querying monthly stats user_id=%d", userID)
	rows, err := r.db.QueryContext(ctx, `
        SELECT
            COALESCE(CAST(strftime('%Y', created_at) AS INTEGER), 0) as year,
            COALESCE(CAST(strftime('%m', created_at) AS INTEGER), 0) as month,
            COALESCE(SUM(CASE WHEN event_type = 'finished_chapter' THEN 1 ELSE 0 END), 0) as chapters_read,
            COALESCE(COUNT(DISTINCT CASE WHEN event_type = 'finished_manga' THEN manga_id END), 0) as manga_completed,
            COALESCE(COUNT(DISTINCT CASE WHEN event_type = 'opened' THEN manga_id END), 0) as manga_started
        FROM reading_history
        WHERE user_id = ?
        GROUP BY year, month
        ORDER BY year DESC, month DESC
        LIMIT 12
    `, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat MonthlyStat
			if err := rows.Scan(&stat.Year, &stat.Month, &stat.ChaptersRead, &stat.MangaCompleted, &stat.MangaStarted); err == nil {
				stats.MonthlyStats = append(stats.MonthlyStats, stat)
			} else {
				log.Printf("history.repository.CalculateReadingStatistics: scan monthly stats failed user_id=%d err=%v", userID, err)
			}
		}
	}

	stats.YearlyStats = []YearlyStat{}
	log.Printf("history.repository.CalculateReadingStatistics: querying yearly stats user_id=%d", userID)
	rows, err = r.db.QueryContext(ctx, `
        SELECT
            COALESCE(CAST(strftime('%Y', created_at) AS INTEGER), 0) as year,
            COALESCE(SUM(CASE WHEN event_type = 'finished_chapter' THEN 1 ELSE 0 END), 0) as chapters_read,
            COALESCE(COUNT(DISTINCT CASE WHEN event_type = 'finished_manga' THEN manga_id END), 0) as manga_completed,
            COALESCE(COUNT(DISTINCT CASE WHEN event_type = 'opened' THEN manga_id END), 0) as manga_started,
            COALESCE(COUNT(DISTINCT date(created_at)), 0) as total_days
        FROM reading_history
        WHERE user_id = ?
        GROUP BY year
        ORDER BY year DESC
        LIMIT 5
    `, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat YearlyStat
			if err := rows.Scan(&stat.Year, &stat.ChaptersRead, &stat.MangaCompleted, &stat.MangaStarted, &stat.TotalDays); err == nil {
				stats.YearlyStats = append(stats.YearlyStats, stat)
			} else {
				log.Printf("history.repository.CalculateReadingStatistics: scan yearly stats failed user_id=%d err=%v", userID, err)
			}
		}
	}

	stats.LastCalculatedAt = time.Now()

	return stats, nil
}

// SaveReadingStatistics persists cached stats
func (r *Repository) SaveReadingStatistics(ctx context.Context, stats *ReadingStatistics) error {
	favoriteGenresJSON, err := json.Marshal(stats.FavoriteGenres)
	if err != nil {
		return err
	}

	monthlyStatsJSON, err := json.Marshal(stats.MonthlyStats)
	if err != nil {
		return err
	}

	yearlyStatsJSON, err := json.Marshal(stats.YearlyStats)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
INSERT INTO reading_statistics (
    user_id, total_chapters_read, total_manga_read, total_manga_reading,
    total_manga_planned, favorite_genres, average_rating, total_reading_time_hours,
    current_streak_days, longest_streak_days, monthly_stats, yearly_stats,
    last_calculated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(user_id) DO UPDATE SET
    total_chapters_read = excluded.total_chapters_read,
    total_manga_read = excluded.total_manga_read,
    total_manga_reading = excluded.total_manga_reading,
    total_manga_planned = excluded.total_manga_planned,
    favorite_genres = excluded.favorite_genres,
    average_rating = excluded.average_rating,
    total_reading_time_hours = excluded.total_reading_time_hours,
    current_streak_days = excluded.current_streak_days,
    longest_streak_days = excluded.longest_streak_days,
    monthly_stats = excluded.monthly_stats,
    yearly_stats = excluded.yearly_stats,
    last_calculated_at = excluded.last_calculated_at
    )
`,
		stats.UserID,
		stats.TotalChaptersRead,
		stats.TotalMangaRead,
		stats.TotalMangaReading,
		stats.TotalMangaPlanned,
		string(favoriteGenresJSON),
		stats.AverageRating,
		stats.TotalReadingTimeHours,
		stats.CurrentStreakDays,
		stats.LongestStreakDays,
		string(monthlyStatsJSON),
		string(yearlyStatsJSON),
		stats.LastCalculatedAt,
	)
	return err
}

// GetCachedReadingStatistics retrieves cached copy
func (r *Repository) GetCachedReadingStatistics(ctx context.Context, userID int64) (*ReadingStatistics, error) {
	query := `
        SELECT
            total_chapters_read,
            total_manga_read,
            total_manga_reading,
            total_manga_planned,
            favorite_genres,
            average_rating,
            total_reading_time_hours,
            current_streak_days,
            longest_streak_days,
            monthly_stats,
            yearly_stats,
            last_calculated_at
        FROM reading_statistics
        WHERE user_id = ?
    `
	var stats ReadingStatistics
	var favoriteGenresJSON, monthlyStatsJSON, yearlyStatsJSON sql.NullString
	var lastCalculatedAt sql.NullTime
	log.Printf("history.repository.GetCachedReadingStatistics: fetching cache for user_id=%d", userID)
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalChaptersRead,
		&stats.TotalMangaRead,
		&stats.TotalMangaReading,
		&stats.TotalMangaPlanned,
		&favoriteGenresJSON,
		&stats.AverageRating,
		&stats.TotalReadingTimeHours,
		&stats.CurrentStreakDays,
		&stats.LongestStreakDays,
		&monthlyStatsJSON,
		&yearlyStatsJSON,
		&lastCalculatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	stats.UserID = userID

	if favoriteGenresJSON.Valid {
		_ = json.Unmarshal([]byte(favoriteGenresJSON.String), &stats.FavoriteGenres)
	}
	if monthlyStatsJSON.Valid {
		_ = json.Unmarshal([]byte(monthlyStatsJSON.String), &stats.MonthlyStats)
	}
	if yearlyStatsJSON.Valid {
		_ = json.Unmarshal([]byte(yearlyStatsJSON.String), &stats.YearlyStats)
	}
	if lastCalculatedAt.Valid {
		stats.LastCalculatedAt = lastCalculatedAt.Time
	}

	return &stats, nil
}

// UpdateReadingGoalProgress updates goal progress values
func (r *Repository) UpdateReadingGoalProgress(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE reading_goals
        SET current_value = (
            SELECT COALESCE(COUNT(*), 0)
            FROM reading_history rh
            WHERE rh.user_id = reading_goals.user_id AND rh.event_type = 'finished_chapter'
        ),
        updated_at = CURRENT_TIMESTAMP
        WHERE user_id = ?
    `, userID)
	return err
}

// GetActiveReadingGoals retrieves goals still active
func (r *Repository) GetActiveReadingGoals(ctx context.Context, userID int64) ([]ReadingGoal, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, user_id, goal_type, target_value, current_value, period_start, period_end, status, created_at, updated_at
        FROM reading_goals
        WHERE user_id = ? AND status = 'active'
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []ReadingGoal
	for rows.Next() {
		var goal ReadingGoal
		var status string
		if err := rows.Scan(
			&goal.GoalID,
			&goal.UserID,
			&goal.GoalType,
			&goal.TargetValue,
			&goal.CurrentValue,
			&goal.StartDate,
			&goal.EndDate,
			&status,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		); err == nil {
			goal.Completed = status == "completed"
			goals = append(goals, goal)
		}
	}
	return goals, rows.Err()
}
