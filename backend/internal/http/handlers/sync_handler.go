package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/internal/tcp"
)

// SyncLayerStatus describes a basic sync status response.
type SyncLayerStatus struct {
	OK      bool      `json:"ok"`
	Message string    `json:"message,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

// TCPSyncStatus captures TCP layer information.
type TCPSyncStatus struct {
	OK      bool   `json:"ok"`
	Devices int    `json:"devices"`
	Message string `json:"message,omitempty"`
}

// CloudSyncStatus captures cloud sync information.
type CloudSyncStatus struct {
	OK       bool      `json:"ok"`
	Message  string    `json:"message,omitempty"`
	LastSync time.Time `json:"last_sync,omitempty"`
	Pending  int       `json:"pending,omitempty"`
}

// SyncStatus represents sync freshness for each layer.
type SyncStatus struct {
	Local SyncLayerStatus `json:"local"`
	TCP   TCPSyncStatus   `json:"tcp"`
	Cloud CloudSyncStatus `json:"cloud"`
}

// SyncStatusHandler exposes live synchronization status backed by runtime data.
type SyncStatusHandler struct {
	db        *sql.DB
	monitor   *db.HealthMonitor
	tcpServer *tcp.Server
	dsn       string
}

// NewSyncStatusHandler builds a new SyncStatusHandler.
func NewSyncStatusHandler(dbConn *sql.DB, monitor *db.HealthMonitor, tcpServer *tcp.Server, dsn string) *SyncStatusHandler {
	return &SyncStatusHandler{
		db:        dbConn,
		monitor:   monitor,
		tcpServer: tcpServer,
		dsn:       dsn,
	}
}

// GetStatus returns synchronization status based on live system state.
func (h *SyncStatusHandler) GetStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := SyncStatus{
		Local: h.collectLocalStatus(ctx),
		TCP:   h.collectTCPStatus(),
		Cloud: h.collectCloudStatus(),
	}

	c.JSON(http.StatusOK, status)
}

func (h *SyncStatusHandler) collectLocalStatus(ctx context.Context) SyncLayerStatus {
	if h.db == nil {
		return SyncLayerStatus{OK: false, Message: "database connection is not configured"}
	}

	checkCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	err := h.db.PingContext(checkCtx)
	healthy := err == nil
	if h.monitor != nil {
		healthy = healthy && h.monitor.IsHealthy()
	}

	message := "database reachable"
	if err != nil {
		message = fmt.Sprintf("database ping failed: %v", err)
	} else if h.monitor != nil && !h.monitor.IsHealthy() {
		message = "database connection is unhealthy"
	}

	updated := time.Time{}
	if h.monitor != nil {
		updated = h.monitor.LastCheck()
	}

	if updated.IsZero() {
		updated = time.Now().UTC()
	}

	return SyncLayerStatus{
		OK:      healthy,
		Message: message,
		Updated: updated,
	}
}

func (h *SyncStatusHandler) collectTCPStatus() TCPSyncStatus {
	if h.tcpServer == nil {
		return TCPSyncStatus{OK: false, Devices: 0, Message: "TCP sync server disabled"}
	}

	stats := h.tcpServer.Stats()
	status := TCPSyncStatus{
		OK:      stats.Running,
		Devices: stats.Clients,
	}

	switch {
	case !stats.Running:
		status.Message = "TCP sync server is not accepting connections"
	case stats.MaxClients > 0:
		status.Message = fmt.Sprintf("%d/%d clients connected", stats.Clients, stats.MaxClients)
	default:
		status.Message = fmt.Sprintf("%d clients connected", stats.Clients)
	}

	return status
}

func (h *SyncStatusHandler) collectCloudStatus() CloudSyncStatus {
	status := CloudSyncStatus{}

	path := h.databaseFilePath()
	if path == "" {
		status.OK = false
		status.Message = "database file path unavailable"
		return status
	}

	info, err := os.Stat(path)
	if err != nil {
		status.OK = false
		status.Message = fmt.Sprintf("backup metadata unavailable: %v", err)
		return status
	}

	status.LastSync = info.ModTime().UTC()
	status.OK = true
	status.Message = "last backup detected"
	return status
}

func (h *SyncStatusHandler) databaseFilePath() string {
	if h.dsn == "" {
		return ""
	}

	dsn := strings.TrimPrefix(h.dsn, "file:")
	if idx := strings.IndexRune(dsn, '?'); idx != -1 {
		dsn = dsn[:idx]
	}

	if dsn == "" {
		return ""
	}

	return dsn
}
