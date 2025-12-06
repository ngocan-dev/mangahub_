package udp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

// AdminHandler xử lý admin commands
type AdminHandler struct {
	server     *Server
	adminToken string // Load từ ENV
}

// AdminCommand represents an admin command packet
type AdminCommand struct {
	Command   string                 `json:"command"`
	Token     string                 `json:"admin_token"`
	Params    map[string]interface{} `json:"params"`
	RequestID string                 `json:"request_id"`
}

// AdminResponse represents admin command response
type AdminResponse struct {
	Success   bool        `json:"success"`
	Command   string      `json:"command"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id"`
}

// ClientInfo thông tin client cho admin
type ClientInfo struct {
	UserID       int64   `json:"user_id"`
	Address      string  `json:"address"`
	DeviceID     string  `json:"device_id"`
	NovelIDs     []int64 `json:"novel_ids"`
	AllNovels    bool    `json:"all_novels"`
	RegisteredAt string  `json:"registered_at"`
	LastSeen     string  `json:"last_seen"`
	Duration     string  `json:"duration"`
}

// NewAdminHandler tạo admin handler
func NewAdminHandler(server *Server) *AdminHandler {
	// Load admin token từ environment
	token := os.Getenv("UDP_ADMIN_TOKEN")
	if token == "" {
		log.Println("WARNING: UDP_ADMIN_TOKEN not set, admin commands disabled")
	}

	return &AdminHandler{
		server:     server,
		adminToken: token,
	}
}

// Handle xử lý admin command packet
func (h *AdminHandler) Handle(ctx context.Context, packet *Packet, addr *net.UDPAddr) {
	// Parse command
	payloadBytes, err := json.Marshal(packet.Payload)
	if err != nil {
		h.sendError(addr, "", "invalid command format")
		return
	}

	var cmd AdminCommand
	if err := json.Unmarshal(payloadBytes, &cmd); err != nil {
		h.sendError(addr, "", "invalid command structure")
		return
	}

	// Authenticate
	if !h.authenticate(cmd.Token) {
		log.Printf("Unauthorized admin command attempt from %s", addr.String())
		h.sendError(addr, cmd.RequestID, "unauthorized: invalid admin token")
		return
	}

	// Execute command
	log.Printf("Admin command received: %s from %s", cmd.Command, addr.String())

	var response interface{}
	var cmdErr error

	switch cmd.Command {
	case "stats":
		response = h.getStats(ctx)
	case "list-clients":
		response = h.listClients(ctx, cmd.Params)
	case "disconnect":
		response = h.disconnectClient(ctx, cmd.Params)
	case "reload-config":
		response = h.reloadConfig(ctx)
	case "health":
		response = h.getHealth(ctx)
	default:
		h.sendError(addr, cmd.RequestID, fmt.Sprintf("unknown command: %s", cmd.Command))
		return
	}

	// Send response
	h.sendResponse(addr, cmd.RequestID, cmd.Command, response, cmdErr)
}

// authenticate kiểm tra admin token
func (h *AdminHandler) authenticate(token string) bool {
	if h.adminToken == "" {
		return false // Admin commands disabled
	}
	return token == h.adminToken
}

// getStats trả về server statistics
func (h *AdminHandler) getStats(ctx context.Context) map[string]interface{} {
	h.server.mu.RLock()
	activeClients := len(h.server.clients)
	clientsByUser := len(h.server.clientsByUser)
	novelCount := len(h.server.clientsByNovel)
	h.server.mu.RUnlock()

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get metrics from metrics system
	metricsData := h.server.metrics.GetStats()

	stats := map[string]interface{}{
		"server": map[string]interface{}{
			"address": h.server.address,
		},
		"clients": map[string]interface{}{
			"active_total":   activeClients,
			"unique_users":   clientsByUser,
			"novels_tracked": novelCount,
		},
		"metrics": metricsData,
		"memory": map[string]interface{}{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"num_cpu":    runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
	}

	return stats
}

// listClients liệt kê clients
func (h *AdminHandler) listClients(ctx context.Context, params map[string]interface{}) interface{} {
	h.server.mu.RLock()
	defer h.server.mu.RUnlock()

	var clients []*ClientInfo

	// Filter by novel_id nếu có
	novelIDFilter := int64(0)
	if novelID, ok := params["novel_id"]; ok {
		if idFloat, ok := novelID.(float64); ok {
			novelIDFilter = int64(idFloat)
		}
	}

	// Filter by user_id nếu có
	userIDFilter := int64(0)
	if userID, ok := params["user_id"]; ok {
		if idFloat, ok := userID.(float64); ok {
			userIDFilter = int64(idFloat)
		}
	}

	for _, client := range h.server.clients {
		// Apply filters
		if novelIDFilter > 0 && !client.IsSubscribedTo(novelIDFilter) {
			continue
		}
		if userIDFilter > 0 && client.UserID != userIDFilter {
			continue
		}

		client.mu.RLock()
		info := &ClientInfo{
			UserID:       client.UserID,
			Address:      client.Address.String(),
			DeviceID:     client.DeviceID,
			NovelIDs:     client.NovelIDs,
			AllNovels:    client.AllNovels,
			RegisteredAt: client.RegisteredAt.Format(time.RFC3339),
			LastSeen:     client.LastSeen.Format(time.RFC3339),
			Duration:     time.Since(client.RegisteredAt).String(),
		}
		client.mu.RUnlock()

		clients = append(clients, info)
	}

	return map[string]interface{}{
		"count":   len(clients),
		"clients": clients,
	}
}

// disconnectClient force disconnect một client
func (h *AdminHandler) disconnectClient(ctx context.Context, params map[string]interface{}) interface{} {
	// Parse user_id from params
	userIDFloat, ok := params["user_id"].(float64)
	if !ok {
		return map[string]string{
			"error": "user_id parameter required",
		}
	}
	userID := int64(userIDFloat)

	// Optional: device_id để disconnect specific device
	deviceID := ""
	if devID, ok := params["device_id"].(string); ok {
		deviceID = devID
	}

	h.server.mu.Lock()
	defer h.server.mu.Unlock()

	disconnected := 0
	var disconnectedClients []string

	// Find and remove clients
	for key, client := range h.server.clients {
		if client.UserID == userID {
			// If device_id specified, only disconnect that device
			if deviceID != "" && client.DeviceID != deviceID {
				continue
			}

			// Send disconnect notification to client
			disconnectPacket := &Packet{
				Type:  PacketTypeError,
				Error: "disconnected by admin",
				Payload: map[string]string{
					"reason": "administrative action",
				},
			}
			h.server.sendPacket(client.Address, disconnectPacket)

			// Remove from registry
			h.server.removeClient(key)

			disconnected++
			disconnectedClients = append(disconnectedClients, key)
		}
	}

	return map[string]interface{}{
		"success":              disconnected > 0,
		"disconnected_count":   disconnected,
		"disconnected_clients": disconnectedClients,
	}
}

// reloadConfig reload server configuration
func (h *AdminHandler) reloadConfig(ctx context.Context) interface{} {
	// Example: Reload configuration from environment or config file
	// This is a placeholder - implement based on your config system

	log.Println("Configuration reload requested by admin")

	return map[string]interface{}{
		"success": true,
		"message": "configuration reload acknowledged",
		"note":    "implement specific reload logic based on your config system",
	}
}

// getHealth trả về health status
func (h *AdminHandler) getHealth(ctx context.Context) interface{} {
	checks := map[string]bool{
		"database":    h.server.db != nil,
		"connection":  h.server.conn != nil,
		"has_clients": h.server.GetClientCount() > 0,
	}

	// Test database connection if exists
	if h.server.db != nil {
		err := h.server.db.PingContext(ctx)
		checks["database"] = err == nil
	}

	// Determine overall status
	status := "healthy"
	for name, ok := range checks {
		if !ok && name != "has_clients" {
			status = "unhealthy"
			break
		}
	}

	if status == "healthy" && !checks["has_clients"] {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// sendResponse gửi response về admin client
func (h *AdminHandler) sendResponse(addr *net.UDPAddr, requestID, command string, data interface{}, err error) {
	response := &Packet{
		Type: PacketTypeAdminResponse,
		Payload: AdminResponse{
			Success:   err == nil,
			Command:   command,
			Data:      data,
			Error:     "",
			RequestID: requestID,
		},
	}

	if err != nil {
		responsePayload := response.Payload.(AdminResponse)
		responsePayload.Error = err.Error()
		response.Payload = responsePayload
	}

	h.server.sendPacket(addr, response)
}

// sendError gửi error response
func (h *AdminHandler) sendError(addr *net.UDPAddr, requestID, message string) {
	h.sendResponse(addr, requestID, "", nil, fmt.Errorf(message))
}
