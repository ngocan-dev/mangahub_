package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gin-gonic/gin"
	"github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/internal/queue"
	"github.com/ngocan-dev/mangahub/backend/internal/tcp"
	"github.com/ngocan-dev/mangahub/backend/internal/udp"
	mangapb "github.com/ngocan-dev/mangahub/backend/proto/manga"
)

// ServiceStatus describes the status of an individual service.
type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Address string `json:"address"`
	Uptime  string `json:"uptime"`
	Load    string `json:"load"`
}

// DatabaseStatus captures connectivity and metadata for the DB.
type DatabaseStatus struct {
	Connection string   `json:"connection"`
	Size       string   `json:"size"`
	Tables     []string `json:"tables"`
	LastBackup string   `json:"last_backup"`
}

// ResourceStatus represents runtime resource usage.
type ResourceStatus struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
	Disk   string `json:"disk"`
}

// ServerStatus is the full payload returned to the CLI.
type ServerStatus struct {
	Overall   string          `json:"overall"`
	Services  []ServiceStatus `json:"services"`
	Database  DatabaseStatus  `json:"database"`
	Resources ResourceStatus  `json:"resources"`
	Issues    []string        `json:"issues"`
}

// StatusHandler exposes the server status endpoint backed by real runtime data.
type StatusHandler struct {
	startTime   time.Time
	db          *sql.DB
	dbHealth    *db.HealthMonitor
	writeQueue  *queue.WriteQueue
	tcpServer   *tcp.Server
	udpServer   *udp.Server
	dsn         string
	apiAddress  string
	grpcAddress string
	tcpAddress  string
	udpAddress  string
	wsAddress   string
}

// NewStatusHandler builds a new StatusHandler with the required dependencies.
func NewStatusHandler(startTime time.Time, dbConn *sql.DB, monitor *db.HealthMonitor, writeQueue *queue.WriteQueue, dsn string) *StatusHandler {
	return &StatusHandler{
		startTime:  startTime,
		db:         dbConn,
		dbHealth:   monitor,
		writeQueue: writeQueue,
		dsn:        dsn,
	}
}

// SetTCPServer wires the TCP sync server instance.
func (h *StatusHandler) SetTCPServer(server *tcp.Server) {
	h.tcpServer = server
}

// SetUDPServer wires the UDP notification server instance.
func (h *StatusHandler) SetUDPServer(server *udp.Server) {
	h.udpServer = server
}

// SetAddresses configures advertised service addresses.
func (h *StatusHandler) SetAddresses(api, grpcAddr, tcpAddr, udpAddr string) {
	h.apiAddress = api
	h.grpcAddress = grpcAddr
	h.tcpAddress = tcpAddr
	h.udpAddress = udpAddr
}

// SetWSAddress configures the WebSocket chat server address.
func (h *StatusHandler) SetWSAddress(addr string) {
	h.wsAddress = addr
}

// GetStatus aggregates live status information for the CLI.
func (h *StatusHandler) GetStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	services, serviceIssues := h.collectServices(ctx)
	dbStatus, dbIssues := h.collectDatabase(ctx)
	resources := collectResources()

	issues := append(serviceIssues, dbIssues...)
	overall := "healthy"
	if len(issues) > 0 {
		overall = "degraded"
	}

	status := ServerStatus{
		Overall:   overall,
		Services:  services,
		Database:  dbStatus,
		Resources: resources,
		Issues:    issues,
	}

	c.JSON(http.StatusOK, status)
}

func (h *StatusHandler) collectServices(ctx context.Context) ([]ServiceStatus, []string) {
	uptime := time.Since(h.startTime).Round(time.Second)
	queueSize := 0
	if h.writeQueue != nil {
		queueSize = h.writeQueue.Size()
	}

	services := []ServiceStatus{{
		Name:    "HTTP API",
		Status:  "online",
		Address: h.apiAddress,
		Uptime:  uptime.String(),
		Load:    fmt.Sprintf("%d queued writes", queueSize),
	}}

	issues := make([]string, 0)

	if h.grpcAddress != "" {
		status, load, err := checkGRPC(ctx, h.grpcAddress)
		if err != nil {
			issues = append(issues, fmt.Sprintf("gRPC server unreachable: %v", err))
		}

		services = append(services, ServiceStatus{
			Name:    "gRPC",
			Status:  status,
			Address: h.grpcAddress,
			Uptime:  uptime.String(),
			Load:    load,
		})
	}

	if h.tcpServer != nil {
		stats := h.tcpServer.Stats()
		svcStatus := "offline"
		load := "not running"

		if stats.Running {
			svcStatus = "online"
			if stats.MaxClients > 0 {
				load = fmt.Sprintf("%d/%d clients", stats.Clients, stats.MaxClients)
				if stats.Clients >= stats.MaxClients {
					issues = append(issues, "TCP sync server at capacity")
				}
			} else {
				load = fmt.Sprintf("%d clients", stats.Clients)
			}
		} else {
			issues = append(issues, "TCP sync server is not accepting connections")
		}

		services = append(services, ServiceStatus{
			Name:    "TCP Sync",
			Status:  svcStatus,
			Address: h.tcpAddress,
			Uptime:  uptime.String(),
			Load:    load,
		})
	}

	if h.udpServer != nil {
		stats := h.udpServer.Stats()
		udpStatus := "offline"
		udpLoad := "not running"

		if stats.Running {
			udpStatus = "online"
			if stats.MaxClients > 0 {
				udpLoad = fmt.Sprintf("%d/%d clients", stats.Clients, stats.MaxClients)
				if stats.Clients >= stats.MaxClients {
					issues = append(issues, "UDP notification server at capacity")
				}
			} else {
				udpLoad = fmt.Sprintf("%d clients", stats.Clients)
			}
		} else {
			issues = append(issues, "UDP notification server is not accepting packets")
		}

		services = append(services, ServiceStatus{
			Name:    "UDP Notifications",
			Status:  udpStatus,
			Address: h.udpAddress,
			Uptime:  uptime.String(),
			Load:    udpLoad,
		})
	}

	if h.wsAddress != "" {
		status, load, err := checkWebSocket(ctx, h.wsAddress)
		if err != nil {
			issues = append(issues, fmt.Sprintf("WebSocket chat server unreachable: %v", err))
		}

		services = append(services, ServiceStatus{
			Name:    "WebSocket Chat",
			Status:  status,
			Address: h.wsAddress,
			Uptime:  uptime.String(),
			Load:    load,
		})
	}

	return services, issues
}

func (h *StatusHandler) collectDatabase(ctx context.Context) (DatabaseStatus, []string) {
	status := DatabaseStatus{}
	issues := make([]string, 0)

	if h.db == nil {
		status.Connection = "error"
		issues = append(issues, "database connection is not configured")
		return status, issues
	}

	pingCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if err := h.db.PingContext(pingCtx); err == nil {
		status.Connection = "active"
	} else {
		status.Connection = "error"
		issues = append(issues, fmt.Sprintf("database ping failed: %v", err))
	}

	if h.dbHealth != nil && !h.dbHealth.IsHealthy() {
		status.Connection = "error"
		issues = append(issues, "database connection is unhealthy")
	}

	size, lastBackup, err := h.databaseStats(ctx)
	if err == nil {
		status.Size = size
		status.LastBackup = lastBackup
	} else {
		issues = append(issues, fmt.Sprintf("database stats unavailable: %v", err))
	}

	tables, err := h.listTables(ctx)
	if err == nil {
		status.Tables = tables
	} else {
		issues = append(issues, fmt.Sprintf("database tables unavailable: %v", err))
	}

	return status, issues
}

func (h *StatusHandler) databaseStats(ctx context.Context) (string, string, error) {
	if h.db == nil {
		return "", "", fmt.Errorf("database not configured")
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var pageCount, pageSize int64
	if err := h.db.QueryRowContext(queryCtx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return "", "", err
	}
	if err := h.db.QueryRowContext(queryCtx, "PRAGMA page_size").Scan(&pageSize); err != nil {
		return "", "", err
	}

	size := formatBytes(pageCount * pageSize)
	lastBackup := h.databaseFileModTime()

	return size, lastBackup, nil
}

func (h *StatusHandler) databaseFilePath() string {
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

	return filepath.Clean(dsn)
}

func (h *StatusHandler) databaseFileModTime() string {
	path := h.databaseFilePath()
	if path == "" {
		return ""
	}

	info, err := os.Stat(path)
	if err != nil {
		return ""
	}

	return info.ModTime().UTC().Format(time.RFC3339)
}

func (h *StatusHandler) listTables(ctx context.Context) ([]string, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	rows, err := h.db.QueryContext(queryCtx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}

func collectResources() ResourceStatus {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	disk := ""
	if usage, err := diskUsage("."); err == nil {
		disk = usage
	}

	return ResourceStatus{
		Memory: fmt.Sprintf("%s used", formatBytes(int64(mem.Alloc))),
		CPU:    fmt.Sprintf("%d goroutines", runtime.NumGoroutine()),
		Disk:   disk,
	}
}

func diskUsage(path string) (string, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return "", err
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bfree) * int64(stat.Bsize)
	used := total - free
	percent := float64(used) / float64(total) * 100

	return fmt.Sprintf("%.1f%% of %s used", percent, formatBytes(total)), nil
}

func checkGRPC(ctx context.Context, address string) (string, string, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return "offline", "unreachable", err
	}
	defer conn.Close()

	client := mangapb.NewMangaServiceClient(conn)
	checkCtx, checkCancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer checkCancel()

	resp, err := client.SearchManga(checkCtx, &mangapb.SearchMangaRequest{Limit: 1})
	if err != nil {
		return "offline", "unreachable", err
	}

	loadParts := []string{}
	if resp.Total > 0 {
		loadParts = append(loadParts, fmt.Sprintf("%d total results", resp.Total))
	}
	if len(resp.Results) > 0 {
		loadParts = append(loadParts, fmt.Sprintf("sample: %s", resp.Results[0].Title))
	}
	if resp.Pages > 0 {
		loadParts = append(loadParts, fmt.Sprintf("page %d/%d", resp.Page, resp.Pages))
	}

	load := "no manga records"
	if len(loadParts) > 0 {
		load = strings.Join(loadParts, " | ")
	}

	return "online", load, nil
}

func checkWebSocket(ctx context.Context, address string) (string, string, error) {
	url := address
	switch {
	case strings.HasPrefix(address, "ws://"):
		url = "http://" + strings.TrimPrefix(address, "ws://")
	case strings.HasPrefix(address, "wss://"):
		url = "https://" + strings.TrimPrefix(address, "wss://")
	case !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://"):
		url = "http://" + strings.TrimPrefix(address, "//")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(url, "/")+"/status", nil)
	if err != nil {
		return "offline", "unreachable", err
	}

	client := &http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return "offline", "unreachable", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "offline", fmt.Sprintf("status %d", resp.StatusCode), fmt.Errorf("websocket status returned %d", resp.StatusCode)
	}

	var wsStatus struct {
		Running bool   `json:"running"`
		Clients int    `json:"clients"`
		Rooms   int    `json:"rooms"`
		Uptime  string `json:"uptime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wsStatus); err != nil {
		return "offline", "invalid status payload", err
	}

	status := "offline"
	if wsStatus.Running {
		status = "online"
	}

	load := fmt.Sprintf("%d clients, %d rooms", wsStatus.Clients, wsStatus.Rooms)
	if wsStatus.Uptime != "" {
		load = fmt.Sprintf("%s | %s", wsStatus.Uptime, load)
	}

	return status, load, nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
