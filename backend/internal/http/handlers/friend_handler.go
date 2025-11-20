package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/manga-backend/domain/friend"
)

// FriendHandler manages friend workflows (search, request, accept)
type FriendHandler struct {
	service *friend.Service
}

// NewFriendHandler constructs a FriendHandler
func NewFriendHandler(service *friend.Service) *FriendHandler {
	return &FriendHandler{service: service}
}

// Search allows a user to look up another user by username
func (h *FriendHandler) Search(c *gin.Context) {
	username := strings.TrimSpace(c.Query("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username query is required"})
		return
	}

	result, err := h.service.SearchUser(c.Request.Context(), username)
	if err != nil {
		if errors.Is(err, friend.ErrInvalidUsername) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
			return
		}
		if errors.Is(err, friend.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search user"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SendRequest sends a friend request to another user
func (h *FriendHandler) SendRequest(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	requesterUsername, _ := c.Get("username")
	usernameString, _ := requesterUsername.(string)

	friendship, err := h.service.SendFriendRequest(c.Request.Context(), userID, usernameString, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, friend.ErrInvalidUsername):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
			return
		case errors.Is(err, friend.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, friend.ErrCannotFriendSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add yourself"})
		case errors.Is(err, friend.ErrAlreadyFriends):
			c.JSON(http.StatusConflict, gin.H{"error": "already friends"})
		case errors.Is(err, friend.ErrRequestPending):
			c.JSON(http.StatusConflict, gin.H{"error": "friend request already pending"})
		case errors.Is(err, friend.ErrBlocked):
			c.JSON(http.StatusForbidden, gin.H{"error": "friendship blocked"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send friend request"})
		}
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// AcceptRequest approves an incoming friend request
func (h *FriendHandler) AcceptRequest(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	var req struct {
		RequesterUsername string `json:"requester_username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requester_username is required"})
		return
	}

	accepterUsername, _ := c.Get("username")
	accepterUsernameStr, _ := accepterUsername.(string)

	friendship, err := h.service.AcceptFriendRequest(c.Request.Context(), userID, accepterUsernameStr, req.RequesterUsername)
	if err != nil {
		switch {
		case errors.Is(err, friend.ErrInvalidUsername):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
			return
		case errors.Is(err, friend.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "requester not found"})
		case errors.Is(err, friend.ErrNoPendingRequest):
			c.JSON(http.StatusNotFound, gin.H{"error": "no pending friend request"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept friend request"})
		}
		return
	}

	c.JSON(http.StatusOK, friendship)
}

// RequireUserID is a helper to extract user ID and fail fast when missing
func RequireUserID(c *gin.Context) (int64, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return 0, false
	}

	switch v := userIDInterface.(type) {
	case int64:
		return v, true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
			return 0, false
		}
		return parsed, true
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return 0, false
	}
}
