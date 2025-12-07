package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Address string `json:"address"`
	Uptime  string `json:"uptime"`
	Load    string `json:"load"`
}

type DatabaseStatus struct {
	Connection string   `json:"connection"`
	Size       string   `json:"size"`
	Tables     []string `json:"tables"`
	LastBackup string   `json:"last_backup"`
}

type ResourceStatus struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
	Disk   string `json:"disk"`
}

type ServerStatus struct {
	Overall   string          `json:"overall"`
	Services  []ServiceStatus `json:"services"`
	Database  DatabaseStatus  `json:"database"`
	Resources ResourceStatus  `json:"resources"`
	Issues    []string        `json:"issues"`
}

func (c *Client) GetServerStatus(ctx context.Context) (*ServerStatus, error) {
	var status ServerStatus
	if err := c.doRequest(ctx, http.MethodGet, "/server/status", nil, &status); err != nil {
		return nil, wrapServerStatusError(err)
	}
	return &status, nil
}

func wrapServerStatusError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("✗ Server status request timed out.\nTry increasing timeout or check network connectivity.")
	}

	var netErr net.Error
	switch {
	case errors.As(err, &netErr) && netErr.Timeout():
		return errors.New("✗ Server status request timed out.\nTry increasing timeout or check network connectivity.")
	case errors.As(err, &netErr):
		return errors.New("✗ Failed to reach MangaHub server.\nCheck if the server is running: mangahub server start")
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
		return errors.New("✗ Invalid response from server.\nPlease update your server or CLI.")
	}

	return fmt.Errorf("✗ Failed to fetch server status\nError: %v\nRun: mangahub server start", err)
}
