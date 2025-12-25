package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/backend/domain/comment"
	"github.com/ngocan-dev/mangahub/backend/domain/history"
	domainlibrary "github.com/ngocan-dev/mangahub/backend/domain/library"
	"github.com/ngocan-dev/mangahub/backend/domain/manga"
	"github.com/ngocan-dev/mangahub/backend/internal/cache"
	"github.com/ngocan-dev/mangahub/backend/internal/queue"
	chapterrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/chapter"
	libraryrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/library"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
	libraryservice "github.com/ngocan-dev/mangahub/backend/internal/service/library"
)

// MangaHandler handles manga-related HTTP endpoints.
type MangaHandler struct {
	DB             *sql.DB
	mangaService   *manga.Service
	libraryService *libraryservice.Service
	historyService *history.Service
	reviewService  *comment.Service
	broadcaster    history.Broadcaster
	dbHealth       manga.DBHealthChecker
	writeQueue     *queue.WriteQueue
}

// GetMangaService builds a manga service with optional cache support.
func GetMangaService(db *sql.DB, mangaCache *cache.MangaCache) *manga.Service {
	mangaService := manga.NewService(db)
	if mangaCache != nil {
		mangaService.SetCache(mangaCache)
	}

	chapterRepo := chapterrepository.NewRepository(db)
	chapterSvc := chapterservice.NewService(chapterRepo)
	mangaService.SetChapterService(chapterSvc)

	return mangaService
}

// NewMangaHandlerWithService constructs a MangaHandler using the provided manga service.
func NewMangaHandlerWithService(db *sql.DB, mangaService *manga.Service) *MangaHandler {
	chapterRepo := chapterrepository.NewRepository(db)
	chapterSvc := chapterservice.NewService(chapterRepo)
	mangaService.SetChapterService(chapterSvc)

	libraryRepo := libraryrepository.NewRepository(db)
	librarySvc := libraryservice.NewService(libraryRepo, mangaService, nil)

	historyRepo := history.NewRepository(db)
	historySvc := history.NewService(historyRepo, chapterSvc, librarySvc, mangaService)

	reviewRepo := comment.NewRepository(db)
	reviewSvc := comment.NewService(reviewRepo, mangaService, nil)

	return &MangaHandler{
		DB:             db,
		mangaService:   mangaService,
		libraryService: librarySvc,
		historyService: historySvc,
		reviewService:  reviewSvc,
	}
}

// SetBroadcaster configures the broadcaster for history updates.
func (h *MangaHandler) SetBroadcaster(b history.Broadcaster) {
	h.broadcaster = b
	if h.historyService != nil {
		h.historyService.SetBroadcaster(b)
	}
}

// SetDBHealth sets the DB health checker on the manga service.
func (h *MangaHandler) SetDBHealth(checker manga.DBHealthChecker) {
	h.dbHealth = checker
	if h.mangaService != nil {
		h.mangaService.SetDBHealth(checker)
	}
}

// SetWriteQueue attaches a write queue to the manga service.
func (h *MangaHandler) SetWriteQueue(q *queue.WriteQueue) {
	h.writeQueue = q
	if h.mangaService != nil {
		h.mangaService.SetWriteQueue(q)
	}
}

// GetPopularManga returns the popular manga list, leveraging cache when available.
func (h *MangaHandler) GetPopularManga(c *gin.Context) {
	limitParam := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitParam)

	popular, err := h.mangaService.GetPopularManga(c.Request.Context(), limit)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, manga.ErrDatabaseUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, popular)
}

// Search finds manga using query parameters.
func (h *MangaHandler) Search(c *gin.Context) {
	var req manga.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search parameters"})
		return
	}

	// Support comma-separated genres
	if len(req.Genres) == 1 && strings.Contains(req.Genres[0], ",") {
		req.Genres = strings.Split(req.Genres[0], ",")
	}

	resp, err := h.mangaService.Search(c.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, manga.ErrDatabaseUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDetails retrieves manga detail information.
func (h *MangaHandler) GetDetails(c *gin.Context) {
	idParam := c.Param("id")
	mangaID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || mangaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var userID *int64
	if val, exists := c.Get("user_id"); exists {
		switch v := val.(type) {
		case int64:
			userID = &v
		case string:
			if parsed, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil {
				userID = &parsed
			}
		}
	}

	detail, err := h.mangaService.GetDetails(c.Request.Context(), mangaID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, manga.ErrMangaNotFound):
			status = http.StatusNotFound
		case errors.Is(err, manga.ErrDatabaseUnavailable):
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	if userID != nil {
		progress, _ := h.historyService.GetProgress(c.Request.Context(), *userID, mangaID)
		detail.UserProgress = progress
		status, _ := h.libraryService.GetLibraryStatus(c.Request.Context(), *userID, mangaID)
		detail.LibraryStatus = status
	}

	c.JSON(http.StatusOK, detail)
}

// GetLibrary lists the authenticated user's library entries.
func (h *MangaHandler) GetLibrary(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	resp, err := h.libraryService.GetLibrary(c.Request.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, libraryservice.ErrDatabaseError) {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	if resp == nil {
		resp = &domainlibrary.GetLibraryResponse{Entries: []domainlibrary.LibraryEntry{}}
	}

	c.JSON(http.StatusOK, resp)
}

// AddToLibrary adds a manga to the authenticated user's library.
func (h *MangaHandler) AddToLibrary(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	mangaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || mangaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req domainlibrary.AddToLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.libraryService.AddToLibrary(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, libraryservice.ErrInvalidStatus):
			status = http.StatusBadRequest
		case errors.Is(err, libraryservice.ErrMangaNotFound):
			status = http.StatusNotFound
		case errors.Is(err, libraryservice.ErrMangaAlreadyInLibrary):
			status = http.StatusConflict
		case errors.Is(err, libraryservice.ErrDatabaseError):
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateProgress updates reading progress for a manga.
func (h *MangaHandler) UpdateProgress(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	mangaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || mangaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req history.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current_chapter is required"})
		return
	}

	resp, err := h.historyService.UpdateProgress(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, history.ErrInvalidChapterNumber):
			status = http.StatusBadRequest
		case errors.Is(err, history.ErrMangaNotFound):
			status = http.StatusNotFound
		case errors.Is(err, history.ErrMangaNotInLibrary):
			status = http.StatusForbidden
		case errors.Is(err, history.ErrDatabaseError):
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateReview creates a review for a manga.
func (h *MangaHandler) CreateReview(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	mangaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || mangaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req comment.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review payload"})
		return
	}

	resp, err := h.reviewService.CreateReview(c.Request.Context(), userID, mangaID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, comment.ErrInvalidReviewRating), errors.Is(err, comment.ErrReviewContentTooShort), errors.Is(err, comment.ErrReviewContentTooLong):
			status = http.StatusBadRequest
		case errors.Is(err, comment.ErrMangaNotFound):
			status = http.StatusNotFound
		case errors.Is(err, comment.ErrMangaNotCompleted):
			status = http.StatusForbidden
		case errors.Is(err, comment.ErrReviewAlreadyExists):
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetReviews returns reviews for a manga.
func (h *MangaHandler) GetReviews(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || mangaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort_by", "recent")

	resp, err := h.reviewService.GetReviews(c.Request.Context(), mangaID, page, limit, sortBy)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, comment.ErrDatabaseError) {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetFriendsActivityFeed lists friend activities for the authenticated user.
func (h *MangaHandler) GetFriendsActivityFeed(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	log.Printf("handler.GetFriendsActivityFeed: user_id=%d page=%d limit=%d", userID, page, limit)
	resp, err := h.historyService.GetFriendsActivityFeed(c.Request.Context(), userID, page, limit)
	if err != nil {
		log.Printf("handler.GetFriendsActivityFeed: user_id=%d error=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load friends activity"})
		return
	}

	if resp == nil {
		resp = &history.ActivityFeedResponse{
			Activities: []history.Activity{},
			Total:      0,
			Page:       page,
			Limit:      limit,
			Pages:      0,
		}
	} else if resp.Activities == nil {
		resp.Activities = []history.Activity{}
	}

	c.JSON(http.StatusOK, resp)
}

// GetReadingStatistics returns reading statistics for the authenticated user.
func (h *MangaHandler) GetReadingStatistics(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		log.Printf("handler.GetReadingStatistics: missing user_id in context")
		return
	}

	force := c.Query("force") == "true"
	log.Printf("handler.GetReadingStatistics: user_id=%d force=%t", userID, force)
	summary, err := h.historyService.GetReadingSummary(c.Request.Context(), userID)
	if err != nil {
		log.Printf("handler.GetReadingStatistics: user_id=%d error=%v", userID, err)
		status := http.StatusInternalServerError
		if errors.Is(err, history.ErrDatabaseError) {
			c.JSON(status, gin.H{"error": "unable to load reading statistics"})
			return
		}
		c.JSON(http.StatusOK, &history.ReadingSummary{})
		return
	}

	if summary == nil {
		summary = &history.ReadingSummary{}
	}

	c.JSON(http.StatusOK, summary)
}

// GetReadingAnalytics filters reading statistics with query params.
func (h *MangaHandler) GetReadingAnalytics(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		log.Printf("handler.GetReadingAnalytics: missing user_id in context")
		return
	}

	var req history.ReadingAnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("handler.GetReadingAnalytics: bind error user_id=%d err=%v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analytics parameters"})
		return
	}

	log.Printf("handler.GetReadingAnalytics: user_id=%d time_period=%s", userID, req.TimePeriod)
	analytics, err := h.historyService.GetReadingAnalyticsBuckets(c.Request.Context(), userID)
	if err != nil {
		log.Printf("handler.GetReadingAnalytics: user_id=%d error=%v", userID, err)
		if errors.Is(err, history.ErrDatabaseError) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load reading analytics"})
			return
		}
		c.JSON(http.StatusOK, &history.ReadingAnalyticsResponse{
			Daily:   []history.ReadingAnalyticsPoint{},
			Weekly:  []history.ReadingAnalyticsPoint{},
			Monthly: []history.ReadingAnalyticsPoint{},
		})
		return
	}

	if analytics == nil {
		analytics = &history.ReadingAnalyticsResponse{
			Daily:   []history.ReadingAnalyticsPoint{},
			Weekly:  []history.ReadingAnalyticsPoint{},
			Monthly: []history.ReadingAnalyticsPoint{},
		}
	} else {
		if analytics.Daily == nil {
			analytics.Daily = []history.ReadingAnalyticsPoint{}
		}
		if analytics.Weekly == nil {
			analytics.Weekly = []history.ReadingAnalyticsPoint{}
		}
		if analytics.Monthly == nil {
			analytics.Monthly = []history.ReadingAnalyticsPoint{}
		}
	}

	c.JSON(http.StatusOK, analytics)
}
