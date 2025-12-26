package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/backend/domain/chat"
)

// ChatMessageHandler exposes REST endpoints for friend-only chat.
type ChatMessageHandler struct {
	service *chat.Service
}

// NewChatMessageHandler builds a handler.
func NewChatMessageHandler(service *chat.Service) *ChatMessageHandler {
	return &ChatMessageHandler{service: service}
}

// SendMessage handles POST /chat/messages.
func (h *ChatMessageHandler) SendMessage(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	var req struct {
		FriendUserID int64  `json:"friend_user_id" binding:"required"`
		Content      string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "friend_user_id and content are required"})
		return
	}

	room, message, err := h.service.SendMessage(c.Request.Context(), userID, req.FriendUserID, req.Content)
	if err != nil {
		switch err {
		case chat.ErrEmptyMessage, chat.ErrCannotMessageSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case chat.ErrNotFriends:
			c.JSON(http.StatusForbidden, gin.H{"error": "friendship not accepted"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"room":    room,
		"message": message,
	})
}

// ListMessages handles GET /chat/rooms/:roomID/messages.
func (h *ChatMessageHandler) ListMessages(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	roomIDParam := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDParam, 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	limit := 100
	offset := 0
	if v := c.Query("limit"); v != "" {
		if parsed, parseErr := strconv.Atoi(v); parseErr == nil {
			limit = parsed
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, parseErr := strconv.Atoi(v); parseErr == nil {
			offset = parsed
		}
	}

	messages, err := h.service.ListMessages(c.Request.Context(), userID, roomID, limit, offset)
	if err != nil {
		switch err {
		case chat.ErrInvalidRoom:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		case chat.ErrNotFriends:
			c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ListConversations handles GET /chat/conversations.
func (h *ChatMessageHandler) ListConversations(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	convs, err := h.service.ListConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load conversations"})
		return
	}

	if convs == nil {
		convs = []chat.ConversationSummary{}
	}

	c.JSON(http.StatusOK, convs)
}
