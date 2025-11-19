package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/auth"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/chapter"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/domain/manga"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/udp"
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

// NotifyChapterRelease handles administrator-triggered chapter release notifications
// Main Success Scenario:
// 1. Administrator triggers notification for specific manga
// 2. System creates notification message with manga details
// 3. UDP server broadcasts message to all registered clients
// 4. Clients receive notification and display to users
// 5. System logs successful broadcast
func (h *NotificationHandler) NotifyChapterRelease(c *gin.Context) {
	// Check authentication (admin required)
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

	// Validate token
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

	// Check if user is admin (basic check - can be enhanced with role-based auth)
	isAdmin, err := h.checkAdminRole(c.Request.Context(), claims.UserID)
	if err != nil {
		log.Printf("Error checking admin role: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin access required",
		})
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
	chapterService := chapter.NewService(chapter.NewRepository(h.DB))
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
		// Try to find chapter ID from chapter number
		valid, foundChapterID, err := chapterService.ValidateChapterNumber(c.Request.Context(), req.NovelID, req.Chapter)
		if err == nil && valid && foundChapterID != nil {
			chapterID = *foundChapterID
		}
		// If not found, we'll proceed with chapterID = 0 (optional field)
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
		"message":    "notification broadcasted successfully",
		"novel_id":   req.NovelID,
		"novel_name": novelName,
		"chapter":    req.Chapter,
	})
}

// checkAdminRole checks if user has admin role
func (h *NotificationHandler) checkAdminRole(ctx context.Context, userID int64) (bool, error) {
	var roleName string
	err := h.DB.QueryRowContext(ctx, `
		SELECT r.RoleName 
		FROM Users u
		LEFT JOIN Roles r ON u.RoleId = r.RoleId
		WHERE u.UserId = ?
	`, userID).Scan(&roleName)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return roleName == "Admin", nil
}
