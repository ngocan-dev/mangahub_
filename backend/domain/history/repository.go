package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

// GetReadingSummary returns lean statistics for a user.
func (r *Repository) GetReadingSummary(ctx context.Context, userID int64) (*ReadingSummary, error) {
	query := `
WITH rp AS (
    SELECT
        COALESCE(SUM(Current_Chapter), 0) AS total_chapters_read,
        COALESCE(COUNT(DISTINCT Novel_Id), 0) AS total_manga,
        MAX(Last_Read_At) AS last_read_at
    FROM Reading_Progress
    WHERE User_Id = ?
),
rs AS (
    SELECT COALESCE(Current_Streak_Days, 0) AS reading_streak
    FROM Reading_Statistics
    WHERE User_Id = ?
)
SELECT
    COALESCE(rp.total_manga, 0) AS total_manga,
    COALESCE(rp.total_chapters_read, 0) AS total_chapters_read,
    COALESCE((SELECT reading_streak FROM rs LIMIT 1), 0) AS reading_streak,
    rp.last_read_at
FROM rp
LIMIT 1
`
	var summary ReadingSummary
	var lastReadAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, userID, userID).Scan(
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
            SELECT date(Last_Read_At) AS bucket_date, COALESCE(SUM(Current_Chapter), 0) AS chapters_read
            FROM Reading_Progress
            WHERE User_Id = ?
            GROUP BY bucket_date
            ORDER BY bucket_date DESC
            LIMIT 30
            `,
		},
		{
			dest: &resp.Weekly,
			sql: `
            SELECT date(Last_Read_At, 'weekday 0', '-6 days') AS bucket_date, COALESCE(SUM(Current_Chapter), 0) AS chapters_read
            FROM Reading_Progress
            WHERE User_Id = ?
            GROUP BY bucket_date
            ORDER BY bucket_date DESC
            LIMIT 12
            `,
		},
		{
			dest: &resp.Monthly,
			sql: `
            SELECT strftime('%Y-%m-01', Last_Read_At) AS bucket_date, COALESCE(SUM(Current_Chapter), 0) AS chapters_read
            FROM Reading_Progress
            WHERE User_Id = ?
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
        SELECT Current_Chapter, Current_Chapter_Id, Last_Read_At
        FROM Reading_Progress
        WHERE User_Id = ? AND Novel_Id = ?
    `
	var progress UserProgress
	var chapterID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&progress.CurrentChapter,
		&chapterID,
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
func (r *Repository) UpdateProgress(ctx context.Context, userID, mangaID int64, chapter int, chapterID *int64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE Reading_Progress
SET Current_Chapter = ?, Current_Chapter_Id = ?, Last_Read_At = CURRENT_TIMESTAMP
WHERE User_Id = ? AND Novel_Id = ?
`, chapter, chapterID, userID, mangaID)
	return err
}

// IsMangaCompleted checks whether the user completed the manga in their library
func (r *Repository) IsMangaCompleted(ctx context.Context, userID, mangaID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM User_Library
WHERE User_Id = ? AND Novel_Id = ? AND Status = 'completed'
`, userID, mangaID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFriends retrieves accepted friend IDs
func (r *Repository) GetFriends(ctx context.Context, userID int64) ([]int64, error) {
	query := `
        SELECT Friend_Id FROM Friends WHERE User_Id = ? AND Status = 'accepted'
        UNION
        SELECT User_Id FROM Friends WHERE Friend_Id = ? AND Status = 'accepted'
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

// GetFriendsActivities returns friend feed entries
func (r *Repository) GetFriendsActivities(ctx context.Context, userID int64, page, limit int) ([]Activity, int, error) {
	log.Printf("history.repository.GetFriendsActivities: start user_id=%d page=%d limit=%d", userID, page, limit)
	friendIDs, err := r.GetFriends(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if len(friendIDs) == 0 {
		return []Activity{}, 0, nil
	}

	placeholders := make([]string, len(friendIDs))
	args := make([]interface{}, len(friendIDs))
	for i, id := range friendIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	placeholderStr := strings.Join(placeholders, ",")

	countQuery := fmt.Sprintf(`
        SELECT COUNT(*) FROM (
            SELECT ul.Completed_At as activity_date
            FROM User_Library ul
            WHERE ul.User_Id IN (%s) AND ul.Status = 'completed' AND ul.Completed_At IS NOT NULL

            UNION ALL

            SELECT r.Created_At as activity_date
            FROM Reviews r
            WHERE r.User_Id IN (%s)

            UNION ALL

            SELECT rs.Rating_Date as activity_date
            FROM Rating_System rs
            WHERE rs.User_Id IN (%s)
        )
	`, placeholderStr, placeholderStr, placeholderStr)

	var total int
	countArgs := append(append(args, args...), args...)
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return []Activity{}, 0, nil
		}
		log.Printf("history.repository.GetFriendsActivities: count query failed user_id=%d err=%v", userID, err)
		return nil, 0, err
	}
	if total == 0 {
		return []Activity{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if page > 10000 {
		page = 10000
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
        SELECT
            activity_id,
            user_id,
            username,
            activity_type,
            manga_id,
            manga_name,
            manga_title,
            manga_image,
            rating,
            review_id,
            review_content,
            completed_at,
            created_at
        FROM (
            SELECT
                ul.Library_Id as activity_id,
                ul.User_Id as user_id,
                u.Username as username,
                'completed_manga' as activity_type,
                ul.Novel_Id as manga_id,
                n.Novel_Name as manga_name,
                n.Title as manga_title,
                n.Image as manga_image,
                NULL as rating,
                NULL as review_id,
                NULL as review_content,
                ul.Completed_At as completed_at,
                ul.Completed_At as created_at
            FROM User_Library ul
            JOIN Users u ON ul.User_Id = u.UserId
            JOIN Novels n ON ul.Novel_Id = n.Novel_Id
            WHERE ul.User_Id IN (%s) AND ul.Status = 'completed' AND ul.Completed_At IS NOT NULL

            UNION ALL

            SELECT
                r.Review_Id as activity_id,
                r.User_Id as user_id,
                u.Username as username,
                'review' as activity_type,
                r.Novel_Id as manga_id,
                n.Novel_Name as manga_name,
                n.Title as manga_title,
                n.Image as manga_image,
                r.Rating as rating,
                r.Review_Id as review_id,
                substr(r.Content, 1, 200) as review_content,
                NULL as completed_at,
                r.Created_At as created_at
            FROM Reviews r
            JOIN Users u ON r.User_Id = u.UserId
            JOIN Novels n ON r.Novel_Id = n.Novel_Id
            WHERE r.User_Id IN (%s)

            UNION ALL

            SELECT
                rs.Rating_Id as activity_id,
                rs.User_Id as user_id,
                u.Username as username,
                'rating' as activity_type,
                rs.Novel_Id as manga_id,
                n.Novel_Name as manga_name,
                n.Title as manga_title,
                n.Image as manga_image,
                rs.Rating as rating,
                NULL as review_id,
                NULL as review_content,
                NULL as completed_at,
                rs.Rating_Date as created_at
            FROM Rating_System rs
            JOIN Users u ON rs.User_Id = u.UserId
            JOIN Novels n ON rs.Novel_Id = n.Novel_Id
            WHERE rs.User_Id IN (%s)
        )
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `, placeholderStr, placeholderStr, placeholderStr)

	rows, err := r.db.QueryContext(ctx, query, append(append(append(args, args...), args...), limit, offset)...)
	if err != nil {
		if err == sql.ErrNoRows {
			return []Activity{}, 0, nil
		}
		log.Printf("history.repository.GetFriendsActivities: list query failed user_id=%d err=%v", userID, err)
		return nil, 0, err
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var activity Activity
		var rating sql.NullInt64
		var reviewID sql.NullInt64
		var reviewContent sql.NullString
		var completedAt sql.NullTime
		if err := rows.Scan(
			&activity.ActivityID,
			&activity.UserID,
			&activity.Username,
			&activity.ActivityType,
			&activity.MangaID,
			&activity.MangaName,
			&activity.MangaTitle,
			&activity.MangaImage,
			&rating,
			&reviewID,
			&reviewContent,
			&completedAt,
			&activity.CreatedAt,
		); err != nil {
			continue
		}
		if rating.Valid {
			val := int(rating.Int64)
			activity.Rating = &val
		}
		if reviewID.Valid {
			val := reviewID.Int64
			activity.ReviewID = &val
		}
		if reviewContent.Valid {
			val := reviewContent.String
			activity.ReviewContent = &val
		}
		if completedAt.Valid {
			activity.CompletedAt = &completedAt.Time
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
            COALESCE(SUM(rp.Current_Chapter), 0) as total_chapters_read,
            COUNT(DISTINCT CASE WHEN ul.Status = 'completed' THEN ul.Novel_Id END) as total_manga_read,
            COUNT(DISTINCT CASE WHEN ul.Status = 'reading' THEN ul.Novel_Id END) as total_manga_reading,
            COUNT(DISTINCT CASE WHEN ul.Status = 'plan_to_read' THEN ul.Novel_Id END) as total_manga_planned,
            COALESCE(AVG(rs.Rating), 0) as average_rating
        FROM User_Library ul
        LEFT JOIN Reading_Progress rp ON ul.User_Id = rp.User_Id AND ul.Novel_Id = rp.Novel_Id
        LEFT JOIN Rating_System rs ON ul.User_Id = rs.User_Id AND ul.Novel_Id = rs.Novel_Id
        WHERE ul.User_Id = ?
    `, userID).Scan(
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

	log.Printf("history.repository.CalculateReadingStatistics: querying favorite genres user_id=%d", userID)
	rows, err := r.db.QueryContext(ctx, `
        SELECT n.Genre, COUNT(*) as manga_count, COALESCE(SUM(rp.Current_Chapter), 0) as chapters_read
        FROM User_Library ul
        JOIN Novels n ON ul.Novel_Id = n.Novel_Id
        LEFT JOIN Reading_Progress rp ON ul.User_Id = rp.User_Id AND ul.Novel_Id = rp.Novel_Id
        WHERE ul.User_Id = ?
        GROUP BY n.Genre
        ORDER BY manga_count DESC, chapters_read DESC
        LIMIT 5
    `, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat GenreStat
			if err := rows.Scan(&stat.Genre, &stat.Count, &stat.Chapters); err == nil {
				stats.FavoriteGenres = append(stats.FavoriteGenres, stat)
			}
		}
	}

	stats.MonthlyStats = []MonthlyStat{}
	log.Printf("history.repository.CalculateReadingStatistics: querying monthly stats user_id=%d", userID)
	rows, err = r.db.QueryContext(ctx, `
        SELECT
            COALESCE(CAST(strftime('%Y', Last_Read_At) AS INTEGER), 0) as year,
            COALESCE(CAST(strftime('%m', Last_Read_At) AS INTEGER), 0) as month,
            COALESCE(SUM(Current_Chapter), 0) as chapters_read,
            COALESCE(COUNT(DISTINCT CASE WHEN ul.Status = 'completed' THEN ul.Novel_Id END), 0) as manga_completed,
            COALESCE(COUNT(DISTINCT CASE WHEN ul.Status = 'reading' THEN ul.Novel_Id END), 0) as manga_started
        FROM Reading_Progress rp
        JOIN User_Library ul ON rp.User_Id = ul.User_Id AND rp.Novel_Id = ul.Novel_Id
        WHERE rp.User_Id = ?
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
            COALESCE(CAST(strftime('%Y', Last_Read_At) AS INTEGER), 0) as year,
            COALESCE(SUM(Current_Chapter), 0) as chapters_read,
            COALESCE(COUNT(DISTINCT CASE WHEN ul.Status = 'completed' THEN ul.Novel_Id END), 0) as manga_completed,
            COALESCE(COUNT(DISTINCT CASE WHEN ul.Status = 'reading' THEN ul.Novel_Id END), 0) as manga_started,
            COALESCE(COUNT(DISTINCT date(Last_Read_At)), 0) as total_days
        FROM Reading_Progress rp
        JOIN User_Library ul ON rp.User_Id = ul.User_Id AND rp.Novel_Id = ul.Novel_Id
        WHERE rp.User_Id = ?
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
INSERT INTO Reading_Statistics (
    User_Id, Total_Chapters_Read, Total_Manga_Read, Total_Manga_Reading,
    Total_Manga_Planned, Favorite_Genres, Average_Rating, Total_Reading_Time_Hours,
    Current_Streak_Days, Longest_Streak_Days, Monthly_Stats, Yearly_Stats,
    Last_Calculated_At
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(User_Id) DO UPDATE SET
    Total_Chapters_Read = excluded.Total_Chapters_Read,
    Total_Manga_Read = excluded.Total_Manga_Read,
    Total_Manga_Reading = excluded.Total_Manga_Reading,
    Total_Manga_Planned = excluded.Total_Manga_Planned,
    Favorite_Genres = excluded.Favorite_Genres,
    Average_Rating = excluded.Average_Rating,
    Total_Reading_Time_Hours = excluded.Total_Reading_Time_Hours,
    Current_Streak_Days = excluded.Current_Streak_Days,
    Longest_Streak_Days = excluded.Longest_Streak_Days,
    Monthly_Stats = excluded.Monthly_Stats,
    Yearly_Stats = excluded.Yearly_Stats,
    Last_Calculated_At = excluded.Last_Calculated_At
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
            Total_Chapters_Read,
            Total_Manga_Read,
            Total_Manga_Reading,
            Total_Manga_Planned,
            Favorite_Genres,
            Average_Rating,
            Total_Reading_Time_Hours,
            Current_Streak_Days,
            Longest_Streak_Days,
            Monthly_Stats,
            Yearly_Stats,
            Last_Calculated_At
        FROM Reading_Statistics
        WHERE User_Id = ?
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
        UPDATE Reading_Goals
        SET Current_Value = (
            SELECT COALESCE(SUM(rp.Current_Chapter), 0)
            FROM Reading_Progress rp
            WHERE rp.User_Id = Reading_Goals.User_Id
        ),
        Updated_At = CURRENT_TIMESTAMP
        WHERE User_Id = ?
    `, userID)
	return err
}

// GetActiveReadingGoals retrieves goals still active
func (r *Repository) GetActiveReadingGoals(ctx context.Context, userID int64) ([]ReadingGoal, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT Goal_Id, User_Id, Goal_Type, Target_Value, Current_Value, Start_Date, End_Date, Completed, Created_At, Updated_At
        FROM Reading_Goals
        WHERE User_Id = ? AND Completed = 0
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []ReadingGoal
	for rows.Next() {
		var goal ReadingGoal
		if err := rows.Scan(
			&goal.GoalID,
			&goal.UserID,
			&goal.GoalType,
			&goal.TargetValue,
			&goal.CurrentValue,
			&goal.StartDate,
			&goal.EndDate,
			&goal.Completed,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		); err == nil {
			goals = append(goals, goal)
		}
	}
	return goals, rows.Err()
}
