package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/auth"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/security"
)

type MangaHandler struct {
	service *manga.Service
}

func NewMangaHandler(db *sql.DB, mangaCache interface{}) *MangaHandler {
	service := GetMangaService(db, mangaCache)
	return &MangaHandler{
		service: service,
	}
}

// NewMangaHandlerWithService creates a handler with an existing service
func NewMangaHandlerWithService(service *manga.Service) *MangaHandler {
	return &MangaHandler{
		service: service,
	}
}

// GetMangaService creates or returns a manga service with cache
func GetMangaService(db *sql.DB, mangaCache interface{}) *manga.Service {
	service := manga.NewService(db)

	// Set cache if available
	if mangaCache != nil {
		if cache, ok := mangaCache.(manga.MangaCacher); ok {
			service.SetCache(cache)
		}
	}

	return service
}

// Search handles manga search requests
// Main Success Scenario:
// 1. User opens advanced search interface
// 2. User selects genres, status, rating range, and year filters
// 3. System constructs complex database query
// 4. System applies full-text search on titles and descriptions
// 5. System returns ranked results based on relevance
func (h *MangaHandler) Search(c *gin.Context) {
	var req manga.SearchRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query parameters",
		})
		return
	}

	// Handle multiple genres from query string (e.g., ?genres=action&genres=adventure)
	genres := c.QueryArray("genres")
	if len(genres) > 0 {
		req.Genres = genres
	}

	// Perform search
	response, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		// A2: Database error - System logs error and returns generic message
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during search: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while searching. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// A1: No results found - System displays "no results" message
	if response.Total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "no results found",
			"results": []manga.Manga{},
			"total":   0,
			"page":    req.Page,
			"limit":   req.Limit,
			"pages":   0,
		})
		return
	}

	// Success: Return paginated results
	c.JSON(http.StatusOK, response)
}

// CreateReview handles review creation requests
// Main Success Scenario:
// 1. User navigates to manga and clicks "Write Review"
// 2. User writes review text and assigns rating (1-10)
// 3. System validates review content and rating
// 4. System saves review to database with timestamp
// 5. System displays review on manga page
func (h *MangaHandler) CreateReview(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user context",
		})
		return
	}

	// Get manga ID from URL
	mangaID, err := getMangaIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid manga id",
		})
		return
	}

	// Bind request body
	var req manga.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create review
	response, err := h.service.CreateReview(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		if errors.Is(err, manga.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "manga not found",
			})
			return
		}
		if errors.Is(err, manga.ErrMangaNotCompleted) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "manga must be in your completed list to write a review",
			})
			return
		}
		if errors.Is(err, manga.ErrReviewAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "you have already written a review for this manga",
			})
			return
		}
		if errors.Is(err, manga.ErrInvalidReviewRating) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "rating must be between 1 and 10",
			})
			return
		}
		if errors.Is(err, manga.ErrReviewContentTooShort) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "review content must be at least 10 characters",
			})
			return
		}
		if errors.Is(err, manga.ErrReviewContentTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "review content must not exceed 5000 characters",
			})
			return
		}
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during review creation: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while creating review. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Step 5: Return created review
	c.JSON(http.StatusCreated, response)
}

// GetReviews handles review retrieval requests
// Main Success Scenario:
// 1. User views manga details page
// 2. System retrieves all reviews for the manga
// 3. System calculates average rating from all reviews
// 4. System displays reviews sorted by helpfulness or date
// 5. User can read individual reviews and ratings
func (h *MangaHandler) GetReviews(c *gin.Context) {
	// Get manga ID from URL
	mangaID, err := getMangaIDFromParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid manga id",
		})
		return
	}

	// Get pagination parameters
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Get sort parameter (date or helpfulness)
	sortBy := c.DefaultQuery("sort_by", "date") // Default: date

	// Get reviews
	response, err := h.service.GetReviews(c.Request.Context(), mangaID, page, limit, sortBy)
	if err != nil {
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during review retrieval: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while retrieving reviews. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetFriendsActivityFeed handles friends activity feed requests
// Main Success Scenario:
// 1. User accesses friends activity page
// 2. System retrieves recent activities from friends
// 3. System displays activities (completed manga, reviews, ratings)
// 4. Activities are sorted by recency
// 5. User can click through to view details
func (h *MangaHandler) GetFriendsActivityFeed(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user context",
		})
		return
	}

	// Get pagination parameters
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Get activity feed
	response, err := h.service.GetFriendsActivityFeed(c.Request.Context(), userID, page, limit)
	if err != nil {
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during activity feed retrieval: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while retrieving activity feed. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetReadingStatistics handles reading statistics requests
// Main Success Scenario:
// 1. System analyzes user's reading progress data
// 2. System calculates total chapters read, favorite genres
// 3. System determines reading patterns and trends
// 4. System generates monthly/yearly statistics
// 5. Statistics are cached for performance
func (h *MangaHandler) GetReadingStatistics(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user context",
		})
		return
	}

	// Check if force recalculate is requested
	forceRecalculate := c.Query("recalculate") == "true"

	// Get statistics
	stats, err := h.service.GetReadingStatistics(c.Request.Context(), userID, forceRecalculate)
	if err != nil {
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during statistics calculation: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while calculating statistics. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetReadingAnalytics handles reading analytics requests with filters
// Main Success Scenario:
// 1. User accesses statistics page
// 2. System retrieves cached statistics or generates new ones
// 3. System displays charts and graphs of reading activity
// 4. User can view different time periods and breakdowns
// 5. System shows reading goals progress if set
func (h *MangaHandler) GetReadingAnalytics(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user context",
		})
		return
	}

	// Step 4: Parse query parameters for time period filtering
	var req manga.ReadingAnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// Use defaults if binding fails
		req.TimePeriod = "all_time"
		req.IncludeGoals = true
	}

	// Parse year and month if provided
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			req.Year = &year
		}
	}
	if monthStr := c.Query("month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil && month >= 1 && month <= 12 {
			req.Month = &month
		}
	}

	// Default to including goals if not specified
	if c.Query("include_goals") == "" {
		req.IncludeGoals = true
	}

	// Get analytics
	stats, err := h.service.GetReadingAnalytics(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during analytics retrieval: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while retrieving analytics. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getMangaIDFromParam extracts manga ID from URL parameter
func getMangaIDFromParam(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid manga id: %s", idStr)
	}
	return id, nil
}

// GetDetails handles manga detail requests
// Main Success Scenario:
// 1. User selects manga from search results or direct URL
// 2. System retrieves manga details from database
// 3. System displays title, author, genres, description, chapter count
// 4. System shows user's current progress if logged in
// 5. User can add manga to library or update progress
func (h *MangaHandler) GetDetails(c *gin.Context) {
	// Get manga ID from URL parameter
	mangaIDStr := c.Param("id")
	var mangaID int64
	if _, err := fmt.Sscanf(mangaIDStr, "%d", &mangaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid manga id",
		})
		return
	}

	// Try to get user ID from JWT token (optional - endpoint works without auth)
	var userID *int64
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Extract token from "Bearer <token>"
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			tokenString = authHeader
		}

		if tokenString != "" {
			// Validate token and extract user ID
			claims, err := auth.ValidateToken(tokenString)
			if err == nil && claims != nil {
				userID = &claims.UserID
			}
			// If token is invalid, we just continue without user ID
		}
	}

	// Get manga details
	detail, err := h.service.GetDetails(c.Request.Context(), mangaID, userID)
	if err != nil {
		if errors.Is(err, manga.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "manga not found",
			})
			return
		}

		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during get details: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while retrieving manga details. please try again later",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Success: Return manga details
	c.JSON(http.StatusOK, detail)
}

// AddToLibrary handles adding manga to user's library
// Main Success Scenario:
// 1. User clicks "Add to Library" from manga details
// 2. System presents status options (Reading, Completed, Plan to Read)
// 3. User selects initial status and current chapter
// 4. System creates user_progress record in database
// 5. System confirms addition and updates UI
func (h *MangaHandler) AddToLibrary(c *gin.Context) {
	// Get user ID from JWT token (required for this endpoint)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	// Extract token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		tokenString = authHeader
	}

	// Validate token and get user ID
	// Invalid tokens are rejected
	// Expired tokens trigger reauthentication
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		// Handle different error types
		if errors.Is(err, auth.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "expired token",
				"message": "your session has expired. please login again",
				"code":    "TOKEN_EXPIRED",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "authentication required",
		})
		return
	}
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "token validation failed",
		})
		return
	}
	userID := claims.UserID

	// Get manga ID from URL parameter
	// Invalid data formats are rejected
	mangaIDStr := c.Param("id")
	var mangaID int64
	if _, err := fmt.Sscanf(mangaIDStr, "%d", &mangaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid manga id",
		})
		return
	}

	// Bind request body
	var req manga.AddToLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body. status is required",
		})
		return
	}

	// Validate status input
	// SQL injection attempts are blocked
	if err := security.DetectSQLInjection(req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid status format",
		})
		return
	}

	// Add to library
	response, err := h.service.AddToLibrary(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		// A1: Manga already in library - System offers to update status
		if errors.Is(err, manga.ErrMangaAlreadyInLibrary) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "manga already in library",
				"message": "use update endpoint to change status",
			})
			return
		}

		// Invalid status
		if errors.Is(err, manga.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid status. valid values: plan_to_read, reading, completed, on_hold, dropped",
			})
			return
		}

		// Manga not found
		if errors.Is(err, manga.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "manga not found",
			})
			return
		}

		// A2: Database error - System logs error and shows retry option
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during add to library: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while adding to library. please try again later",
				"retry": true,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Success: Return confirmation
	c.JSON(http.StatusCreated, response)
}

// UpdateProgress handles updating reading progress
// Main Success Scenario:
// 1. User updates current chapter number
// 2. System validates chapter number against manga metadata
// 3. System updates user_progress record with timestamp
// 4. System triggers TCP broadcast to connected clients
// 5. System confirms update to user
func (h *MangaHandler) UpdateProgress(c *gin.Context) {
	// Get user ID from JWT token (required for this endpoint)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	// Extract token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		tokenString = authHeader
	}

	// Validate token and get user ID
	// Invalid tokens are rejected
	// Expired tokens trigger reauthentication
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		// Handle different error types
		if errors.Is(err, auth.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "expired token",
				"message": "your session has expired. please login again",
				"code":    "TOKEN_EXPIRED",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "authentication required",
		})
		return
	}
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "token validation failed",
		})
		return
	}
	userID := claims.UserID

	// Get manga ID from URL parameter
	// Invalid data formats are rejected
	mangaIDStr := c.Param("id")
	var mangaID int64
	if _, err := fmt.Sscanf(mangaIDStr, "%d", &mangaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid manga id",
		})
		return
	}

	// Bind request body
	var req manga.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body. current_chapter is required",
		})
		return
	}

	// Validate chapter number
	// Invalid data formats are rejected
	if err := security.ValidatePositiveInteger(req.CurrentChapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid chapter number",
		})
		return
	}

	// Update progress
	response, err := h.service.UpdateProgress(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		// A1: Invalid chapter number - System shows validation error
		if errors.Is(err, manga.ErrInvalidChapterNumber) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Manga not found
		if errors.Is(err, manga.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "manga not found",
			})
			return
		}

		// Manga not in library
		if errors.Is(err, manga.ErrMangaNotInLibrary) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "manga not in library. add it to your library first",
			})
			return
		}

		// Database error
		if errors.Is(err, manga.ErrDatabaseError) {
			log.Printf("Database error during progress update: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "an error occurred while updating progress. please try again later",
			})
			return
		}

		// Check for wrapped errors (invalid chapter wrapped in fmt.Errorf)
		errStr := err.Error()
		if errStr != "" && errors.Is(err, manga.ErrInvalidChapterNumber) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Success: Return confirmation
	c.JSON(http.StatusOK, response)
}
