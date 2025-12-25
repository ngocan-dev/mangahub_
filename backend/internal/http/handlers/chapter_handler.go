package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	chapterrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/chapter"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
	pkgchapter "github.com/ngocan-dev/mangahub/backend/pkg/models"
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

	resp := mapChapterResponse(chapter)

	c.JSON(http.StatusOK, resp)
}

type chapterResponse struct {
	ID            int64  `json:"id"`
	MangaID       int64  `json:"manga_id"`
	ChapterNumber int    `json:"chapter_number"`
	Title         string `json:"title"`
	ContentText   string `json:"content_text"`
	CreatedAt     string `json:"created_at"`
}

func mapChapterResponse(ch *pkgchapter.Chapter) chapterResponse {
	createdAt := ""
	if ch.CreatedAt != nil {
		createdAt = ch.CreatedAt.UTC().Format(time.RFC3339)
	}

	return chapterResponse{
		ID:            ch.ID,
		MangaID:       ch.MangaID,
		ChapterNumber: ch.Number,
		Title:         ch.Title,
		ContentText:   ch.ContentText,
		CreatedAt:     createdAt,
	}
}
