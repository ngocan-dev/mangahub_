package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/ngocan-dev/mangahub/backend/internal/websocket"
)

type ChatHandler struct {
	hub *ws.DirectChatHub
}

func NewChatHandler(hub *ws.DirectChatHub) *ChatHandler {
	return &ChatHandler{hub: hub}
}

func (h *ChatHandler) Serve(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// ✅ ENTRY POINT DUY NHẤT
	h.hub.HandleWS(c.Writer, c.Request, userID)
}
