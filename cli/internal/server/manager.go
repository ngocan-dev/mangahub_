package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
)

type Component struct {
	Key           string
	Name          string
	Address       string
	StartMessages []string
	StatusLabel   string
	Load          string
	Uptime        string
	Indicator     string
	StopLabel     string
}

var (
	stateMu sync.Mutex
	state   = struct {
		running    bool
		degraded   bool
		components []Component
	}{}
)

// Components returns the current component definitions.
func Components() []Component {
	stateMu.Lock()
	defer stateMu.Unlock()
	copy := make([]Component, len(state.components))
	copy = append(copy[:0], state.components...)
	return copy
}

// UpdateFromStatus refreshes component metadata using live server status data.
func UpdateFromStatus(status *api.ServerStatus) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if status == nil {
		state.components = nil
		return
	}

	components := make([]Component, 0, len(status.Services))
	for _, svc := range status.Services {
		key := strings.ToLower(strings.ReplaceAll(svc.Name, " ", "-"))
		components = append(components, Component{
			Key:         key,
			Name:        svc.Name,
			Address:     svc.Address,
			StatusLabel: output.FormatStatus(svc.Status),
			Load:        svc.Load,
			Uptime:      svc.Uptime,
			Indicator:   svc.Name,
			StopLabel:   fmt.Sprintf("%s stopped", svc.Name),
		})
	}

	state.components = components
	state.degraded = strings.EqualFold(status.Overall, "degraded")
}

// MarkRunning records the running components and mode.
func MarkRunning() {
	stateMu.Lock()
	defer stateMu.Unlock()
	state.running = true
}

// MarkStopped marks the server components as stopped.
func MarkStopped() {
	stateMu.Lock()
	defer stateMu.Unlock()
	state.running = false
}

// Running reports whether the server is running.
func Running() bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	return state.running
}

// LogPath returns the location of the server logs.
func LogPath() string {
	return "~/.mangahub/logs/server.log"
}

// SetDegraded toggles degraded mode for status/health commands.
func SetDegraded(value bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	state.degraded = value
}

// Degraded reports whether the system is degraded.
func Degraded() bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	return state.degraded
}
