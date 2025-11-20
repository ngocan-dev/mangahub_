package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/manga-backend/domain/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/auth"
	chapterrepository "github.com/ngocan-dev/mangahub/manga-backend/internal/repository/chapter"
	chapterservice "github.com/ngocan-dev/mangahub/manga-backend/internal/service/chapter"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/udp"
)

type NotificationHandler struct {
	DB       *sql.DB
	notifier *udp.Notifier
}

func NewNotificationHandler(db *sql.DB, notifier *udp.Notifier) *NotificationHandler {
	return &NotificationHandler{
		DB:       db,
		notifier: notifier,
	}
}

// SetNotifier sets the UDP notifier (can be called after UDP server starts)
func (h *NotificationHandler) SetNotifier(notifier *udp.Notifier) {
	h.notifier = notifier
}

// NotifyChapterReleaseRequest represents request to notify about chapter release
type NotifyChapterReleaseRequest struct {
	NovelID   int64 `json:"novel_id" binding:"required"`
	Chapter   int   `json:"chapter" binding:"required"`
	ChapterID int64 `json:"chapter_id,omitempty"`
}

func (h *NotificationHandler) NotifyChapterRelease(c *gin.Context) {
	_, ok := getNotificationClaims(c)
	if !ok {
		return
	}

	// Bind request body
	var req NotifyChapterReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body. novel_id and chapter are required",
		})
		return
	}

	// Step 2: Get manga details
	mangaService := GetMangaService(h.DB, nil)
	chapterService := chapterservice.NewService(chapterrepository.NewRepository(h.DB))
	mangaService.SetChapterService(chapterService)
	mangaDetail, err := mangaService.GetDetails(c.Request.Context(), req.NovelID, nil)
	if err != nil {
		if errors.Is(err, manga.ErrMangaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "manga not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve manga details",
		})
		return
	}

	// Step 2: Create notification message with manga details
	novelName := mangaDetail.Name
	if novelName == "" {
		novelName = mangaDetail.Title
	}

	chapterID := req.ChapterID
	if chapterID == 0 {
		summary, err := chapterService.ValidateChapter(c.Request.Context(), req.NovelID, req.Chapter)
		if err == nil && summary != nil {
			chapterID = summary.ID
		}
	}

	// Step 3: Broadcast via UDP server
	if h.notifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "notification service unavailable",
		})
		return
	}

	err = h.notifier.NotifyChapterRelease(c.Request.Context(), req.NovelID, novelName, req.Chapter, chapterID)
	if err != nil {
		log.Printf("Error broadcasting notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to broadcast notification",
		})
		return
	}

	// Step 5: Return success
	c.JSON(http.StatusOK, gin.H{
		"message":    "chapter release notification sent",
		"novel_id":   req.NovelID,
		"novel_name": novelName,
		"chapter":    req.Chapter,
		"chapter_id": chapterID,
	})
}

func getNotificationClaims(c *gin.Context) (*auth.Claims, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authorization header required",
			"message": "authentication required",
		})
		return nil, false
	}

	tokenString := strings.TrimSpace(authHeader)
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authorization header required",
			"message": "authentication required",
		})
		return nil, false
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "token validation failed",
		})
		return nil, false
	}

	return claims, true
}
