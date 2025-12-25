package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/backend/domain/friend"
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
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}

	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	results, err := h.service.SearchUsers(c.Request.Context(), userID, query)
	if err != nil {
		log.Printf("handler.Search: search error query=%q err=%v", query, err)
		if errors.Is(err, friend.ErrInvalidUsername) {
			c.JSON(http.StatusOK, gin.H{"users": []friend.UserSummary{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search users"})
		return
	}

	if results == nil {
		results = []friend.UserSummary{}
	}

	c.JSON(http.StatusOK, gin.H{"users": results})
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
		TargetUserID int64 `json:"target_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id is required"})
		return
	}

	requesterUsername, _ := c.Get("username")
	usernameString, _ := requesterUsername.(string)

	friendship, err := h.service.SendFriendRequest(c.Request.Context(), userID, usernameString, req.TargetUserID)
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
		RequestID int64 `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	accepterUsername, _ := c.Get("username")
	accepterUsernameStr, _ := accepterUsername.(string)

	friendship, err := h.service.AcceptFriendRequest(c.Request.Context(), userID, accepterUsernameStr, req.RequestID)
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

// RejectRequest rejects a pending friend request
func (h *FriendHandler) RejectRequest(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}
	var req struct {
		RequestID int64 `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}
	if err := h.service.RejectFriendRequest(c.Request.Context(), userID, req.RequestID); err != nil {
		if errors.Is(err, friend.ErrNoPendingRequest) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no pending friend request"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject friend request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "request rejected"})
}

// ListFriends returns accepted friends for current user
func (h *FriendHandler) ListFriends(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}
	friends, err := h.service.ListFriends(c.Request.Context(), userID)
	if err != nil {
		log.Printf("handler.ListFriends: user_id=%d err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load friends"})
		return
	}
	if friends == nil {
		friends = []friend.UserSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

// PendingRequests lists incoming friend requests for the authenticated user.
func (h *FriendHandler) PendingRequests(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}
	reqs, err := h.service.ListPendingRequests(c.Request.Context(), userID)
	if err != nil {
		log.Printf("handler.PendingRequests: user_id=%d err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load friend requests"})
		return
	}
	if reqs == nil {
		reqs = []friend.FriendRequest{}
	}
	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

// RequireUserID is a helper to extract user ID and fail fast when missing
func RequireUserID(c *gin.Context) (int64, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("RequireUserID: missing user_id in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		c.Abort()
		return 0, false
	}

	switch v := userIDInterface.(type) {
	case int64:
		if v <= 0 {
			log.Printf("RequireUserID: non-positive user_id %d", v)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
			c.Abort()
			return 0, false
		}
		return v, true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Printf("RequireUserID: invalid user_id type string=%q err=%v", v, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
			c.Abort()
			return 0, false
		}
		if parsed <= 0 {
			log.Printf("RequireUserID: non-positive user_id string=%q", v)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
			c.Abort()
			return 0, false
		}
		return parsed, true
	default:
		log.Printf("RequireUserID: unsupported user_id type %T", userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		c.Abort()
		return 0, false
	}
}
