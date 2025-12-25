package handlers

import (
	"github.com/gin-gonic/gin"

	ws "github.com/ngocan-dev/mangahub/backend/internal/websocket"
)

// ChatHandler bridges HTTP to the direct friend chat hub.
type ChatHandler struct {
	hub *ws.DirectChatHub
}

// NewChatHandler builds a chat handler
func NewChatHandler(hub *ws.DirectChatHub) *ChatHandler {
	return &ChatHandler{hub: hub}
}

// Serve upgrades HTTP to WebSocket for 1-to-1 friend chat.
func (h *ChatHandler) Serve(c *gin.Context) {
	userID, ok := RequireUserID(c)
	if !ok {
		return
	}
	h.hub.HandleDirectChat(c.Writer, c.Request, userID)
}
