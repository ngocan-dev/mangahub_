package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, validate origin
		return true
	},
}

// ServeWS handles WebSocket requests from clients
// Main Success Scenario:
// 1. User's browser initiates WebSocket connection
// 2. Server upgrades HTTP connection to WebSocket
// 3. Client sends join message with user credentials
// 4. Server validates user and adds to active connections
// 5. Server broadcasts user join notification to other users
// 6. User receives recent chat history
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Step 2: Server upgrades HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := NewClient(hub, conn)

	// Register client (will be authenticated in join message)
	hub.register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
