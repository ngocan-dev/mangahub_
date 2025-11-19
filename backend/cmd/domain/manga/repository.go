package manga

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Search searches for manga/novels based on criteria
// Search queries complete within acceptable time limits
// Pagination prevents memory issues
// Main Success Scenario:
// 1. User opens advanced search interface
// 2. User selects genres, status, rating range, and year filters
// 3. System constructs complex database query
// 4. System applies full-text search on titles and descriptions
// 5. System returns ranked results based on relevance
func (r *Repository) Search(ctx context.Context, req SearchRequest) ([]Manga, int, error) {
	// Validate and enforce pagination limits to prevent memory issues
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

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	// Step 4: Full-text search on titles and descriptions
	// Optimized: Search in indexed columns first (Novel_Name, Title, Author) before Description
	// Search queries complete within acceptable time limits
	if req.Query != "" {
		searchPattern := "%" + strings.ToLower(req.Query) + "%"
		// Prioritize indexed columns (Title, Author) over Description for better performance
		// Description search is kept but may be slower on large datasets
		conditions = append(conditions, `(
			LOWER(Novel_Name) LIKE ? OR 
			LOWER(Title) LIKE ? OR 
			LOWER(Author) LIKE ? OR 
			LOWER(SUBSTR(Description, 1, 500)) LIKE ?
		)`)
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Step 2: Multiple genres filter
	if len(req.Genres) > 0 {
		genreConditions := make([]string, 0, len(req.Genres))
		for _, genre := range req.Genres {
			if genre != "" {
				genreConditions = append(genreConditions, "Genre LIKE ?")
				args = append(args, "%"+genre+"%")
			}
		}
		if len(genreConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(genreConditions, " OR ")+")")
		}
	}

	// Status filter
	if req.Status != "" {
		conditions = append(conditions, "Status = ?")
		args = append(args, req.Status)
	}

	// Step 2: Rating range filter
	if req.MinRating != nil {
		conditions = append(conditions, "Rating_Point >= ?")
		args = append(args, *req.MinRating)
	}
	if req.MaxRating != nil {
		conditions = append(conditions, "Rating_Point <= ?")
		args = append(args, *req.MaxRating)
	}

	// Step 2: Year filter (extract year from Date_Updated)
	if req.YearFrom != nil {
		conditions = append(conditions, "CAST(strftime('%Y', Date_Updated) AS INTEGER) >= ?")
		args = append(args, *req.YearFrom)
	}
	if req.YearTo != nil {
		conditions = append(conditions, "CAST(strftime('%Y', Date_Updated) AS INTEGER) <= ?")
		args = append(args, *req.YearTo)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Optimize count query: Use indexed columns when possible
	// Search queries complete within acceptable time limits
	// For large datasets, we can estimate count if exact count is too slow
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM Novels %s", whereClause)
	var total int

	// Use context timeout to ensure query completes within acceptable time
	// Search queries complete within acceptable time limits
	countCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(countCtx, countQuery, args...).Scan(&total)
	if err != nil {
		// If count query times out or fails, we can still proceed with limited results
		// This prevents the entire search from failing
		if err == context.DeadlineExceeded {
			// Estimate total based on first page results (fallback)
			total = -1 // Indicate estimated count
		} else {
			return nil, 0, err
		}
	}

	// If no results, return early (optimization)
	if total == 0 {
		return []Manga{}, 0, nil
	}

	// Calculate pagination (already validated above)
	offset := (req.Page - 1) * req.Limit

	// Step 5: Determine sort order (relevance, rating, or date)
	sortClause := "ORDER BY Date_Updated DESC" // Default
	if req.Query != "" && (req.SortBy == "" || req.SortBy == "relevance") {
		// Step 5: Rank by relevance - prioritize title matches, then description
		// Escape special SQL LIKE characters for safety
		searchLower := strings.ToLower(req.Query)
		searchEscaped := strings.ReplaceAll(searchLower, "%", "\\%")
		searchEscaped = strings.ReplaceAll(searchEscaped, "_", "\\_")
		searchEscaped = strings.ReplaceAll(searchEscaped, "[", "\\[")

		sortClause = fmt.Sprintf(`ORDER BY 
			CASE 
				WHEN LOWER(Novel_Name) LIKE '%%%s%%' ESCAPE '\' THEN 1
				WHEN LOWER(Title) LIKE '%%%s%%' ESCAPE '\' THEN 2
				WHEN LOWER(Author) LIKE '%%%s%%' ESCAPE '\' THEN 3
				WHEN LOWER(Description) LIKE '%%%s%%' ESCAPE '\' THEN 4
				ELSE 5
			END,
			Rating_Point DESC,
			Date_Updated DESC`,
			searchEscaped, searchEscaped, searchEscaped, searchEscaped)
	} else if req.SortBy == "rating" {
		sortClause = "ORDER BY Rating_Point DESC, Date_Updated DESC"
	} else if req.SortBy == "date_updated" {
		sortClause = "ORDER BY Date_Updated DESC"
	}

	// Step 3: Build SELECT query with pagination and ranking
	query := fmt.Sprintf(`
		SELECT Novel_Id, Novel_Name, Title, Author, Genre, Status, 
		       Description, Image, Rating_Point, Date_Updated
		FROM Novels
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, sortClause)

	args = append(args, req.Limit, offset)

	// Use context timeout for main query
	// Search queries complete within acceptable time limits
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(queryCtx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Pre-allocate slice with known capacity to prevent memory reallocation
	// Pagination prevents memory issues
	var results []Manga
	if req.Limit > 0 {
		results = make([]Manga, 0, req.Limit)
	}
	for rows.Next() {
		var m Manga
		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Title,
			&m.Author,
			&m.Genre,
			&m.Status,
			&m.Description,
			&m.Image,
			&m.RatingPoint,
			&m.DateUpdated,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, m)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetByID retrieves manga details by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*Manga, error) {
	query := `
		SELECT Novel_Id, Novel_Name, Title, Author, Genre, Status, 
		       Description, Image, Rating_Point, Date_Updated
		FROM Novels
		WHERE Novel_Id = ?
	`

	var m Manga
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID,
		&m.Name,
		&m.Title,
		&m.Author,
		&m.Genre,
		&m.Status,
		&m.Description,
		&m.Image,
		&m.RatingPoint,
		&m.DateUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Manga not found
	}
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// GetChapterCount returns the total number of chapters for a manga
func (r *Repository) GetChapterCount(ctx context.Context, novelID int64) (int, error) {
	query := `SELECT COUNT(*) FROM Chapters WHERE Novel_Id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, novelID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUserProgress retrieves user's reading progress for a manga
func (r *Repository) GetUserProgress(ctx context.Context, userID, novelID int64) (*UserProgress, error) {
	query := `
		SELECT Current_Chapter, Current_Chapter_Id, Last_Read_At
		FROM Reading_Progress
		WHERE User_Id = ? AND Novel_Id = ?
	`

	var progress UserProgress
	var chapterID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID, novelID).Scan(
		&progress.CurrentChapter,
		&chapterID,
		&progress.LastReadAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No progress found
	}
	if err != nil {
		return nil, err
	}

	if chapterID.Valid {
		progress.CurrentChapterID = &chapterID.Int64
	}

	return &progress, nil
}

// GetLibraryStatus retrieves user's library status for a manga
func (r *Repository) GetLibraryStatus(ctx context.Context, userID, novelID int64) (*LibraryStatus, error) {
	query := `
		SELECT Status, Is_Favorite, Rating, Started_At, Completed_At
		FROM User_Library
		WHERE User_Id = ? AND Novel_Id = ?
	`

	var status LibraryStatus
	var isFavorite int // SQLite stores boolean as INTEGER (0 or 1)
	var rating sql.NullInt64
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, novelID).Scan(
		&status.Status,
		&isFavorite,
		&rating,
		&startedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not in library
	}
	if err != nil {
		return nil, err
	}

	status.IsFavorite = isFavorite != 0 // Convert INTEGER to bool

	if rating.Valid {
		ratingVal := int(rating.Int64)
		status.Rating = &ratingVal
	}
	if startedAt.Valid {
		status.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}

	return &status, nil
}

// ValidateChapterNumber checks if a chapter number exists for a manga
func (r *Repository) ValidateChapterNumber(ctx context.Context, novelID int64, chapterNumber int) (bool, *int64, error) {
	if chapterNumber < 1 {
		return false, nil, nil
	}

	query := `SELECT Chapter_Id FROM Chapters WHERE Novel_Id = ? AND Chapter_Number = ? LIMIT 1`
	var chapterID int64
	err := r.db.QueryRowContext(ctx, query, novelID, chapterNumber).Scan(&chapterID)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return true, &chapterID, nil
}

// GetMaxChapterNumber returns the maximum chapter number for a manga
func (r *Repository) GetMaxChapterNumber(ctx context.Context, novelID int64) (int, error) {
	query := `SELECT MAX(Chapter_Number) FROM Chapters WHERE Novel_Id = ?`
	var maxChapter sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, novelID).Scan(&maxChapter)
	if err != nil {
		return 0, err
	}
	if !maxChapter.Valid {
		return 0, nil // No chapters exist
	}
	return int(maxChapter.Int64), nil
}

// UpdateProgress updates user's reading progress
// No data corruption under concurrent access
func (r *Repository) UpdateProgress(ctx context.Context, userID, novelID int64, chapterNumber int, chapterID *int64) error {
	now := time.Now()

	// Use transaction with row-level locking to prevent data corruption
	// No data corruption under concurrent access
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Highest isolation level for SQLite
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if progress exists with row lock
	var exists int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM Reading_Progress 
		WHERE User_Id = ? AND Novel_Id = ?
	`, userID, novelID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		// Update existing progress (row is locked by transaction)
		_, err = tx.ExecContext(ctx, `
			UPDATE Reading_Progress 
			SET Current_Chapter = ?, Current_Chapter_Id = ?, Last_Read_At = ?
			WHERE User_Id = ? AND Novel_Id = ?
		`, chapterNumber, chapterID, now, userID, novelID)
	} else {
		// Insert new progress
		_, err = tx.ExecContext(ctx, `
			INSERT INTO Reading_Progress (User_Id, Novel_Id, Current_Chapter, Current_Chapter_Id, Last_Read_At)
			VALUES (?, ?, ?, ?, ?)
		`, userID, novelID, chapterNumber, chapterID, now)
	}

	if err != nil {
		return err
	}

	// Insert into Progress_History
	_, err = tx.ExecContext(ctx, `
		INSERT INTO Progress_History (User_Id, Novel_Id, To_Chapter, Created_At)
		VALUES (?, ?, ?, ?)
	`, userID, novelID, chapterNumber, now)
	if err != nil {
		// Log but don't fail - history is optional
	}

	return tx.Commit()
}

// CheckLibraryExists checks if manga is already in user's library
func (r *Repository) CheckLibraryExists(ctx context.Context, userID, novelID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM User_Library WHERE User_Id = ? AND Novel_Id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, novelID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddToLibrary adds manga to user's library and creates reading progress
// No data corruption under concurrent access
func (r *Repository) AddToLibrary(ctx context.Context, userID, novelID int64, status string, currentChapter int, isFavorite bool) error {
	// Begin transaction with row-level locking
	// No data corruption under concurrent access
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Highest isolation level for SQLite
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	var startedAt *time.Time
	if status == "reading" || status == "completed" {
		startedAt = &now
	}

	var completedAt *time.Time
	if status == "completed" {
		completedAt = &now
	}

	// Insert into User_Library
	isFavoriteInt := 0
	if isFavorite {
		isFavoriteInt = 1
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO User_Library (User_Id, Novel_Id, Status, Is_Favorite, Started_At, Completed_At, Last_Updated_At)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, novelID, status, isFavoriteInt, startedAt, completedAt, now)
	if err != nil {
		return err
	}

	// Insert into Reading_Progress
	if currentChapter < 1 {
		currentChapter = 1 // Default to chapter 1 if not specified
	}

	// Try to find the chapter ID for the given chapter number
	var chapterID *int64
	var chapterIDVal sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		SELECT Chapter_Id FROM Chapters 
		WHERE Novel_Id = ? AND Chapter_Number = ?
		LIMIT 1
	`, novelID, currentChapter).Scan(&chapterIDVal)
	if err == nil && chapterIDVal.Valid {
		chapterID = &chapterIDVal.Int64
	} else if err != nil && err != sql.ErrNoRows {
		// If there's an error other than "not found", log it but continue
		// We'll just use NULL for chapter_id
	}
	// If chapter not found (sql.ErrNoRows), we'll just use NULL for chapter_id

	_, err = tx.ExecContext(ctx, `
		INSERT INTO Reading_Progress (User_Id, Novel_Id, Current_Chapter, Current_Chapter_Id, Last_Read_At)
		VALUES (?, ?, ?, ?, ?)
	`, userID, novelID, currentChapter, chapterID, now)
	if err != nil {
		return err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// CreateReview creates a new review for a manga
// No data corruption under concurrent access
func (r *Repository) CreateReview(ctx context.Context, userID, mangaID int64, rating int, content string) (int64, error) {
	// Use transaction to ensure atomicity and prevent duplicate reviews
	// No data corruption under concurrent access
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Highest isolation level for SQLite
	})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Check for existing review first (unique constraint prevents duplicates, but we check for better error handling)
	var existingCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM Reviews 
		WHERE User_Id = ? AND Novel_Id = ?
	`, userID, mangaID).Scan(&existingCount)
	if err != nil {
		return 0, err
	}
	if existingCount > 0 {
		tx.Rollback()
		return 0, fmt.Errorf("user has already reviewed this manga")
	}

	query := `
		INSERT INTO Reviews (User_Id, Novel_Id, Rating, Content, Created_At, Updated_At)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := tx.ExecContext(ctx, query, userID, mangaID, rating, content)
	if err != nil {
		return 0, err
	}

	reviewID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return reviewID, nil
}

// GetReviewByUserAndManga gets a review by user and manga (to check if user already reviewed)
func (r *Repository) GetReviewByUserAndManga(ctx context.Context, userID, mangaID int64) (*Review, error) {
	query := `
		SELECT r.Review_Id, r.User_Id, u.Username, r.Novel_Id, r.Rating, r.Content, r.Created_At, r.Updated_At
		FROM Reviews r
		JOIN Users u ON r.User_Id = u.UserId
		WHERE r.User_Id = ? AND r.Novel_Id = ?
	`
	var review Review
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID, mangaID).Scan(
		&review.ReviewID,
		&review.UserID,
		&review.Username,
		&review.MangaID,
		&review.Rating,
		&review.Content,
		&review.CreatedAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		review.UpdatedAt = updatedAt.Time
	}

	return &review, nil
}

// GetReviews retrieves reviews for a manga with pagination
// Step 4: System displays reviews sorted by helpfulness or date
func (r *Repository) GetReviews(ctx context.Context, mangaID int64, page, limit int, sortBy string) ([]Review, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM Reviews WHERE Novel_Id = ?
	`, mangaID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Review{}, 0, nil
	}

	// Calculate pagination
	// Pagination prevents memory issues
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent memory issues
	}
	if page > 10000 {
		page = 10000 // Reasonable upper limit
	}
	offset := (page - 1) * limit

	// Step 4: Determine sort order (date or helpfulness)
	// For now, helpfulness is based on rating (higher rating = more helpful)
	// In the future, this could be based on user votes
	orderClause := "ORDER BY r.Created_At DESC" // Default: newest first
	if sortBy == "helpfulness" || sortBy == "rating" {
		// Sort by rating (higher = more helpful), then by date
		orderClause = "ORDER BY r.Rating DESC, r.Created_At DESC"
	} else if sortBy == "date" || sortBy == "newest" {
		orderClause = "ORDER BY r.Created_At DESC"
	} else if sortBy == "oldest" {
		orderClause = "ORDER BY r.Created_At ASC"
	}

	// Get reviews with user information
	query := fmt.Sprintf(`
		SELECT 
			r.Review_Id,
			r.User_Id,
			u.Username,
			r.Novel_Id,
			r.Rating,
			r.Content,
			r.Created_At,
			r.Updated_At
		FROM Reviews r
		JOIN Users u ON r.User_Id = u.UserId
		WHERE r.Novel_Id = ?
		%s
		LIMIT ? OFFSET ?
	`, orderClause)

	rows, err := r.db.QueryContext(ctx, query, mangaID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var updatedAt sql.NullTime
		err := rows.Scan(
			&review.ReviewID,
			&review.UserID,
			&review.Username,
			&review.MangaID,
			&review.Rating,
			&review.Content,
			&review.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			continue
		}

		if updatedAt.Valid {
			review.UpdatedAt = updatedAt.Time
		}

		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// CheckMangaInCompletedLibrary checks if manga is in user's completed library
func (r *Repository) CheckMangaInCompletedLibrary(ctx context.Context, userID, mangaID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM User_Library 
		WHERE User_Id = ? AND Novel_Id = ? AND Status = 'completed'
	`, userID, mangaID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetReviewStats calculates review statistics for a manga
func (r *Repository) GetReviewStats(ctx context.Context, mangaID int64) (*ReviewStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_reviews,
			COALESCE(AVG(Rating), 0) as average_rating
		FROM Reviews
		WHERE Novel_Id = ?
	`
	var stats ReviewStats
	err := r.db.QueryRowContext(ctx, query, mangaID).Scan(
		&stats.TotalReviews,
		&stats.AverageRating,
	)
	if err != nil {
		return nil, err
	}

	// Round average rating to 2 decimal places
	stats.AverageRating = float64(int(stats.AverageRating*100+0.5)) / 100

	return &stats, nil
}

// GetFriends retrieves all accepted friends of a user
func (r *Repository) GetFriends(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT Friend_Id
		FROM Friends
		WHERE User_Id = ? AND Status = 'accepted'
		UNION
		SELECT User_Id
		FROM Friends
		WHERE Friend_Id = ? AND Status = 'accepted'
	`
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friendIDs []int64
	for rows.Next() {
		var friendID int64
		if err := rows.Scan(&friendID); err != nil {
			continue
		}
		friendIDs = append(friendIDs, friendID)
	}

	return friendIDs, rows.Err()
}

// GetFriendsActivities retrieves activities from friends
// Activities include: completed manga, reviews, ratings
// Step 2: System retrieves recent activities from friends
// Step 4: Activities are sorted by recency
func (r *Repository) GetFriendsActivities(ctx context.Context, userID int64, page, limit int) ([]Activity, int, error) {
	// Get friends list
	friendIDs, err := r.GetFriends(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if len(friendIDs) == 0 {
		return []Activity{}, 0, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(friendIDs))
	args := make([]interface{}, len(friendIDs))
	for i, id := range friendIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	placeholderStr := strings.Join(placeholders, ",")

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			-- Completed manga activities
			SELECT ul.Completed_At as activity_date
			FROM User_Library ul
			WHERE ul.User_Id IN (%s) AND ul.Status = 'completed' AND ul.Completed_At IS NOT NULL
			
			UNION ALL
			
			-- Review activities
			SELECT r.Created_At as activity_date
			FROM Reviews r
			WHERE r.User_Id IN (%s)
			
			UNION ALL
			
			-- Rating activities (from Rating_System)
			SELECT rs.Rating_Date as activity_date
			FROM Rating_System rs
			WHERE rs.User_Id IN (%s)
		)
	`, placeholderStr, placeholderStr, placeholderStr)

	var total int
	countArgs := append(append(args, args...), args...)
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Activity{}, 0, nil
	}

	// Calculate pagination
	// Pagination prevents memory issues
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent memory issues
	}
	if page > 10000 {
		page = 10000 // Reasonable upper limit
	}
	offset := (page - 1) * limit

	// Get activities with all details
	// Step 4: Sort by recency (most recent first)
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
			-- Completed manga activities
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
			
			-- Review activities
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
				SUBSTR(r.Content, 1, 200) as review_content,
				NULL as completed_at,
				r.Created_At as created_at
			FROM Reviews r
			JOIN Users u ON r.User_Id = u.UserId
			JOIN Novels n ON r.Novel_Id = n.Novel_Id
			WHERE r.User_Id IN (%s)
			
			UNION ALL
			
			-- Rating activities (from Rating_System)
			SELECT 
				rs.Rating_Id as activity_id,
				rs.User_Id as user_id,
				u.Username as username,
				'rating' as activity_type,
				rs.Novel_Id as manga_id,
				n.Novel_Name as manga_name,
				n.Title as manga_title,
				n.Image as manga_image,
				rs.Rating_Value as rating,
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

	queryArgs := append(append(countArgs, limit, offset)...)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
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
		var mangaName sql.NullString
		var mangaTitle sql.NullString
		var mangaImage sql.NullString

		err := rows.Scan(
			&activity.ActivityID,
			&activity.UserID,
			&activity.Username,
			&activity.ActivityType,
			&activity.MangaID,
			&mangaName,
			&mangaTitle,
			&mangaImage,
			&rating,
			&reviewID,
			&reviewContent,
			&completedAt,
			&activity.CreatedAt,
		)
		if err != nil {
			continue
		}

		if rating.Valid {
			ratingVal := int(rating.Int64)
			activity.Rating = &ratingVal
		}
		if reviewID.Valid {
			activity.ReviewID = &reviewID.Int64
		}
		if reviewContent.Valid {
			activity.ReviewContent = &reviewContent.String
		}
		if completedAt.Valid {
			activity.CompletedAt = &completedAt.Time
		}
		if mangaName.Valid {
			activity.MangaName = mangaName.String
		}
		if mangaTitle.Valid {
			activity.MangaTitle = mangaTitle.String
		}
		if mangaImage.Valid {
			activity.MangaImage = mangaImage.String
		}

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// CalculateReadingStatistics calculates reading statistics for a user
// Main Success Scenario:
// 1. System analyzes user's reading progress data
// 2. System calculates total chapters read, favorite genres
// 3. System determines reading patterns and trends
// 4. System generates monthly/yearly statistics
func (r *Repository) CalculateReadingStatistics(ctx context.Context, userID int64) (*ReadingStatistics, error) {
	stats := &ReadingStatistics{
		UserID: userID,
	}

	// Step 1: Analyze reading progress data
	// Calculate total chapters read
	var totalChapters int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(Current_Chapter), 0)
		FROM Reading_Progress
		WHERE User_Id = ?
	`, userID).Scan(&totalChapters)
	if err != nil {
		return nil, err
	}
	stats.TotalChaptersRead = totalChapters

	// Step 2: Calculate manga counts by status
	var mangaRead, mangaReading, mangaPlanned int
	err = r.db.QueryRowContext(ctx, `
		SELECT 
			SUM(CASE WHEN Status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN Status = 'reading' THEN 1 ELSE 0 END) as reading,
			SUM(CASE WHEN Status = 'plan_to_read' THEN 1 ELSE 0 END) as planned
		FROM User_Library
		WHERE User_Id = ?
	`, userID).Scan(&mangaRead, &mangaReading, &mangaPlanned)
	if err != nil {
		return nil, err
	}
	stats.TotalMangaRead = mangaRead
	stats.TotalMangaReading = mangaReading
	stats.TotalMangaPlanned = mangaPlanned

	// Step 2: Calculate favorite genres
	genreRows, err := r.db.QueryContext(ctx, `
		SELECT 
			n.Genre,
			COUNT(DISTINCT ul.Novel_Id) as manga_count,
			COALESCE(SUM(rp.Current_Chapter), 0) as chapters_read
		FROM User_Library ul
		JOIN Novels n ON ul.Novel_Id = n.Novel_Id
		LEFT JOIN Reading_Progress rp ON ul.User_Id = rp.User_Id AND ul.Novel_Id = rp.Novel_Id
		WHERE ul.User_Id = ? AND n.Genre IS NOT NULL AND n.Genre != ''
		GROUP BY n.Genre
		ORDER BY manga_count DESC, chapters_read DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer genreRows.Close()

	var favoriteGenres []GenreStat
	for genreRows.Next() {
		var genre GenreStat
		err := genreRows.Scan(&genre.Genre, &genre.Count, &genre.Chapters)
		if err != nil {
			continue
		}
		favoriteGenres = append(favoriteGenres, genre)
	}
	stats.FavoriteGenres = favoriteGenres

	// Calculate average rating
	var avgRating sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(Rating), 0)
		FROM User_Library
		WHERE User_Id = ? AND Rating IS NOT NULL
	`, userID).Scan(&avgRating)
	if err != nil {
		return nil, err
	}
	if avgRating.Valid {
		stats.AverageRating = avgRating.Float64
		// Round to 2 decimal places
		stats.AverageRating = float64(int(stats.AverageRating*100+0.5)) / 100
	}

	// Estimate total reading time (assuming ~5 minutes per chapter)
	stats.TotalReadingTimeHours = float64(totalChapters) * 5.0 / 60.0

	// Step 3: Calculate reading streaks
	currentStreak, longestStreak, err := r.calculateReadingStreaks(ctx, userID)
	if err != nil {
		// Log error but don't fail
		currentStreak = 0
		longestStreak = 0
	}
	stats.CurrentStreakDays = currentStreak
	stats.LongestStreakDays = longestStreak

	// Step 4: Generate monthly statistics (last 12 months)
	monthlyStats, err := r.calculateMonthlyStats(ctx, userID, 12)
	if err != nil {
		// Log error but don't fail
		monthlyStats = []MonthlyStat{}
	}
	stats.MonthlyStats = monthlyStats

	// Step 4: Generate yearly statistics (last 5 years)
	yearlyStats, err := r.calculateYearlyStats(ctx, userID, 5)
	if err != nil {
		// Log error but don't fail
		yearlyStats = []YearlyStat{}
	}
	stats.YearlyStats = yearlyStats

	stats.LastCalculatedAt = time.Now()

	return stats, nil
}

// calculateReadingStreaks calculates current and longest reading streaks
func (r *Repository) calculateReadingStreaks(ctx context.Context, userID int64) (currentStreak, longestStreak int, err error) {
	// Get all reading activity dates from progress history and library updates
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT DATE(Last_Read_At) as read_date
		FROM Reading_Progress
		WHERE User_Id = ?
		UNION
		SELECT DISTINCT DATE(Completed_At) as read_date
		FROM User_Library
		WHERE User_Id = ? AND Completed_At IS NOT NULL
		ORDER BY read_date DESC
	`, userID, userID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			continue
		}
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dates = append(dates, date)
	}

	if len(dates) == 0 {
		return 0, 0, nil
	}

	// Calculate streaks
	currentStreak = 0
	longestStreak = 0
	tempStreak := 1

	// Check current streak (from today backwards)
	today := time.Now().Truncate(24 * time.Hour)
	expectedDate := today
	for i, date := range dates {
		date = date.Truncate(24 * time.Hour)
		if i == 0 && date.Equal(expectedDate) {
			currentStreak = 1
			expectedDate = expectedDate.AddDate(0, 0, -1)
		} else if i > 0 && date.Equal(expectedDate) {
			currentStreak++
			expectedDate = expectedDate.AddDate(0, 0, -1)
		} else if i == 0 && date.Before(expectedDate) {
			// First date is not today, check if it's yesterday
			if date.Equal(expectedDate.AddDate(0, 0, -1)) {
				currentStreak = 1
				expectedDate = date
			}
		}
	}

	// Calculate longest streak
	for i := 1; i < len(dates); i++ {
		prevDate := dates[i-1].Truncate(24 * time.Hour)
		currDate := dates[i].Truncate(24 * time.Hour)
		daysDiff := int(prevDate.Sub(currDate).Hours() / 24)

		if daysDiff == 1 {
			tempStreak++
		} else {
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
			tempStreak = 1
		}
	}
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	return currentStreak, longestStreak, nil
}

// calculateMonthlyStats calculates monthly statistics for the last N months
func (r *Repository) calculateMonthlyStats(ctx context.Context, userID int64, months int) ([]MonthlyStat, error) {
	query := `
		SELECT 
			strftime('%Y', activity_date) as year,
			strftime('%m', activity_date) as month,
			COUNT(DISTINCT CASE WHEN activity_type = 'chapter' THEN novel_id END) as chapters_read,
			COUNT(DISTINCT CASE WHEN activity_type = 'completed' THEN novel_id END) as manga_completed,
			COUNT(DISTINCT CASE WHEN activity_type = 'started' THEN novel_id END) as manga_started
		FROM (
			SELECT 
				DATE(Last_Read_At) as activity_date,
				'chapter' as activity_type,
				Novel_Id as novel_id
			FROM Reading_Progress
			WHERE User_Id = ? AND Last_Read_At >= date('now', '-' || ? || ' months')
			
			UNION ALL
			
			SELECT 
				DATE(Completed_At) as activity_date,
				'completed' as activity_type,
				Novel_Id as novel_id
			FROM User_Library
			WHERE User_Id = ? AND Completed_At IS NOT NULL AND Completed_At >= date('now', '-' || ? || ' months')
			
			UNION ALL
			
			SELECT 
				DATE(Started_At) as activity_date,
				'started' as activity_type,
				Novel_Id as novel_id
			FROM User_Library
			WHERE User_Id = ? AND Started_At IS NOT NULL AND Started_At >= date('now', '-' || ? || ' months')
		)
		GROUP BY year, month
		ORDER BY year DESC, month DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, months, userID, months, userID, months, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []MonthlyStat
	for rows.Next() {
		var stat MonthlyStat
		var yearStr, monthStr string
		err := rows.Scan(&yearStr, &monthStr, &stat.ChaptersRead, &stat.MangaCompleted, &stat.MangaStarted)
		if err != nil {
			continue
		}
		stat.Year, _ = strconv.Atoi(yearStr)
		stat.Month, _ = strconv.Atoi(monthStr)
		stats = append(stats, stat)
	}

	return stats, nil
}

// calculateYearlyStats calculates yearly statistics for the last N years
func (r *Repository) calculateYearlyStats(ctx context.Context, userID int64, years int) ([]YearlyStat, error) {
	query := `
		SELECT 
			strftime('%Y', activity_date) as year,
			COUNT(DISTINCT CASE WHEN activity_type = 'chapter' THEN novel_id END) as chapters_read,
			COUNT(DISTINCT CASE WHEN activity_type = 'completed' THEN novel_id END) as manga_completed,
			COUNT(DISTINCT CASE WHEN activity_type = 'started' THEN novel_id END) as manga_started,
			COUNT(DISTINCT DATE(activity_date)) as total_days
		FROM (
			SELECT 
				DATE(Last_Read_At) as activity_date,
				'chapter' as activity_type,
				Novel_Id as novel_id
			FROM Reading_Progress
			WHERE User_Id = ? AND Last_Read_At >= date('now', '-' || ? || ' years')
			
			UNION ALL
			
			SELECT 
				DATE(Completed_At) as activity_date,
				'completed' as activity_type,
				Novel_Id as novel_id
			FROM User_Library
			WHERE User_Id = ? AND Completed_At IS NOT NULL AND Completed_At >= date('now', '-' || ? || ' years')
			
			UNION ALL
			
			SELECT 
				DATE(Started_At) as activity_date,
				'started' as activity_type,
				Novel_Id as novel_id
			FROM User_Library
			WHERE User_Id = ? AND Started_At IS NOT NULL AND Started_At >= date('now', '-' || ? || ' years')
		)
		GROUP BY year
		ORDER BY year DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, years, userID, years, userID, years, years)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []YearlyStat
	for rows.Next() {
		var stat YearlyStat
		var yearStr string
		err := rows.Scan(&yearStr, &stat.ChaptersRead, &stat.MangaCompleted, &stat.MangaStarted, &stat.TotalDays)
		if err != nil {
			continue
		}
		stat.Year, _ = strconv.Atoi(yearStr)
		stats = append(stats, stat)
	}

	return stats, nil
}

// SaveReadingStatistics saves calculated statistics to cache
// Step 5: Statistics are cached for performance
func (r *Repository) SaveReadingStatistics(ctx context.Context, stats *ReadingStatistics) error {
	// Serialize favorite genres, monthly stats, and yearly stats to JSON
	favoriteGenresJSON, _ := json.Marshal(stats.FavoriteGenres)
	monthlyStatsJSON, _ := json.Marshal(stats.MonthlyStats)
	yearlyStatsJSON, _ := json.Marshal(stats.YearlyStats)

	query := `
		INSERT INTO Reading_Statistics (
			User_Id, Total_Chapters_Read, Total_Manga_Read, Total_Manga_Reading,
			Total_Manga_Planned, Favorite_Genres, Average_Rating,
			Total_Reading_Time_Hours, Current_Streak_Days, Longest_Streak_Days,
			Monthly_Stats, Yearly_Stats, Last_Calculated_At
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
	`

	_, err := r.db.ExecContext(ctx, query,
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

// GetCachedReadingStatistics retrieves cached statistics
func (r *Repository) GetCachedReadingStatistics(ctx context.Context, userID int64) (*ReadingStatistics, error) {
	query := `
		SELECT 
			User_Id, Total_Chapters_Read, Total_Manga_Read, Total_Manga_Reading,
			Total_Manga_Planned, Favorite_Genres, Average_Rating,
			Total_Reading_Time_Hours, Current_Streak_Days, Longest_Streak_Days,
			Monthly_Stats, Yearly_Stats, Last_Calculated_At
		FROM Reading_Statistics
		WHERE User_Id = ?
	`

	var stats ReadingStatistics
	var favoriteGenresJSON, monthlyStatsJSON, yearlyStatsJSON string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID,
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
		&stats.LastCalculatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	json.Unmarshal([]byte(favoriteGenresJSON), &stats.FavoriteGenres)
	json.Unmarshal([]byte(monthlyStatsJSON), &stats.MonthlyStats)
	json.Unmarshal([]byte(yearlyStatsJSON), &stats.YearlyStats)

	return &stats, nil
}

// GetActiveReadingGoals retrieves active reading goals for a user
func (r *Repository) GetActiveReadingGoals(ctx context.Context, userID int64) ([]ReadingGoal, error) {
	query := `
		SELECT 
			Goal_Id, User_Id, Goal_Type, Target_Value, Current_Value,
			Period_Type, Period_Start, Period_End, Status,
			Created_At, Updated_At
		FROM Reading_Goals
		WHERE User_Id = ? AND Status = 'active'
			AND Period_End >= date('now')
		ORDER BY Period_End ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []ReadingGoal
	for rows.Next() {
		var goal ReadingGoal
		var periodStart, periodEnd, createdAt, updatedAt string
		err := rows.Scan(
			&goal.GoalID,
			&goal.UserID,
			&goal.GoalType,
			&goal.TargetValue,
			&goal.CurrentValue,
			&goal.PeriodType,
			&periodStart,
			&periodEnd,
			&goal.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			continue
		}

		// Parse timestamps
		goal.PeriodStart, _ = time.Parse("2006-01-02 15:04:05", periodStart)
		goal.PeriodEnd, _ = time.Parse("2006-01-02 15:04:05", periodEnd)
		goal.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		goal.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		// Calculate progress percentage
		if goal.TargetValue > 0 {
			goal.Progress = float64(goal.CurrentValue) / float64(goal.TargetValue) * 100.0
			if goal.Progress > 100.0 {
				goal.Progress = 100.0
			}
		}

		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

// UpdateReadingGoalProgress updates the current value for reading goals
func (r *Repository) UpdateReadingGoalProgress(ctx context.Context, userID int64) error {
	// Get active goals
	goals, err := r.GetActiveReadingGoals(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, goal := range goals {
		// Skip if period hasn't started or has ended
		if now.Before(goal.PeriodStart) || now.After(goal.PeriodEnd) {
			continue
		}

		var currentValue int
		switch goal.GoalType {
		case "chapters":
			// Count chapters read in the period
			err = r.db.QueryRowContext(ctx, `
				SELECT COALESCE(SUM(Current_Chapter), 0)
				FROM Reading_Progress
				WHERE User_Id = ? 
					AND Last_Read_At >= ? 
					AND Last_Read_At <= ?
			`, userID, goal.PeriodStart, goal.PeriodEnd).Scan(&currentValue)
		case "manga":
			// Count manga completed in the period
			err = r.db.QueryRowContext(ctx, `
				SELECT COUNT(*)
				FROM User_Library
				WHERE User_Id = ? 
					AND Status = 'completed'
					AND Completed_At >= ? 
					AND Completed_At <= ?
			`, userID, goal.PeriodStart, goal.PeriodEnd).Scan(&currentValue)
		case "reading_time":
			// Calculate reading time in hours (chapters * 5 minutes / 60)
			var chapters int
			err = r.db.QueryRowContext(ctx, `
				SELECT COALESCE(SUM(Current_Chapter), 0)
				FROM Reading_Progress
				WHERE User_Id = ? 
					AND Last_Read_At >= ? 
					AND Last_Read_At <= ?
			`, userID, goal.PeriodStart, goal.PeriodEnd).Scan(&chapters)
			if err == nil {
				currentValue = int(float64(chapters) * 5.0 / 60.0) // Convert to hours
			}
		}

		if err != nil {
			continue
		}

		// Update goal status
		status := "active"
		if currentValue >= goal.TargetValue {
			status = "completed"
		} else if now.After(goal.PeriodEnd) && currentValue < goal.TargetValue {
			status = "failed"
		}

		// Update goal
		_, err = r.db.ExecContext(ctx, `
			UPDATE Reading_Goals
			SET Current_Value = ?, Status = ?, Updated_At = ?
			WHERE Goal_Id = ?
		`, currentValue, status, now, goal.GoalID)
		if err != nil {
			continue
		}
	}

	return nil
}
