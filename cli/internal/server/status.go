package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
)

// FetchStatus retrieves the live server status from the backend API.
func FetchStatus(ctx context.Context, client *api.Client) (*api.ServerStatus, error) {
	if client == nil {
		return nil, fmt.Errorf("api client is required")
	}

	status, err := client.GetServerStatus(ctx)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// FilterByMode returns a copy of the server status filtered by the selected service flags.
func FilterByMode(status *api.ServerStatus, httpOnly, tcpOnly, udpOnly bool) *api.ServerStatus {
	if status == nil {
		return nil
	}

	requested := []struct {
		enabled bool
		match   func(api.ServiceStatus) bool
	}{
		{httpOnly, func(s api.ServiceStatus) bool { return strings.Contains(strings.ToLower(s.Name), "http") }},
		{tcpOnly, func(s api.ServiceStatus) bool { return strings.Contains(strings.ToLower(s.Name), "tcp") }},
		{udpOnly, func(s api.ServiceStatus) bool { return strings.Contains(strings.ToLower(s.Name), "udp") }},
	}

	applyFilter := false
	for _, req := range requested {
		if req.enabled {
			applyFilter = true
			break
		}
	}
	if !applyFilter {
		return status
	}

	filtered := *status
	filtered.Services = filtered.Services[:0]

	for _, svc := range status.Services {
		for _, req := range requested {
			if req.enabled && req.match(svc) {
				filtered.Services = append(filtered.Services, svc)
				break
			}
		}
	}

	return &filtered
}
