package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	chapterrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/chapter"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
)

// ChapterHandler handles chapter-specific endpoints.
type ChapterHandler struct {
	DB *sql.DB
}

// NewChapterHandler constructs a ChapterHandler.
func NewChapterHandler(db *sql.DB) *ChapterHandler {
	return &ChapterHandler{DB: db}
}

// GetChapter retrieves a chapter by its identifier.
func (h *ChapterHandler) GetChapter(c *gin.Context) {
	chapterID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || chapterID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	chapterSvc := chapterservice.NewService(chapterrepository.NewRepository(h.DB))
	chapter, err := chapterSvc.GetChapterByID(c.Request.Context(), chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if chapter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}
