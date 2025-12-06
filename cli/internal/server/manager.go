package server

import "sync"

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
	}{components: defaultComponents()}
)

// defaultComponents returns the predefined MangaHub server components.
func defaultComponents() []Component {
	return []Component{
		{
			Key:     "http",
			Name:    "HTTP API Server",
			Address: "http://localhost:8080",
			StartMessages: []string{
				"      ✓ Starting on http://localhost:8080",
				"      ✓ Database connection established",
				"      ✓ JWT middleware loaded",
				"      ✓ 12 routes registered",
				"      Status: Running",
			},
			StatusLabel: "✓ Online",
			Load:        "12 req/min",
			Uptime:      "2h 15m",
			Indicator:   "HTTP API",
			StopLabel:   "HTTP API stopped",
		},
		{
			Key:     "tcp",
			Name:    "TCP Sync Server",
			Address: "tcp://localhost:9090",
			StartMessages: []string{
				"      ✓ Starting on tcp://localhost:9090",
				"      ✓ Connection pool initialized (max: 100)",
				"      ✓ Broadcast channels ready",
				"      Status: Listening for connections",
			},
			StatusLabel: "✓ Online",
			Load:        "3 clients",
			Uptime:      "2h 15m",
			Indicator:   "TCP Sync",
			StopLabel:   "TCP Sync stopped",
		},
		{
			Key:     "udp",
			Name:    "UDP Notification Server",
			Address: "udp://localhost:9091",
			StartMessages: []string{
				"      ✓ Starting on udp://localhost:9091",
				"      ✓ Client registry initialized",
				"      ✓ Notification queue ready",
				"      Status: Ready for broadcasts",
			},
			StatusLabel: "✓ Online",
			Load:        "8 clients",
			Uptime:      "2h 15m",
			Indicator:   "UDP Notifications",
			StopLabel:   "UDP Notifications stopped",
		},
		{
			Key:     "grpc",
			Name:    "gRPC Internal Service",
			Address: "grpc://localhost:9092",
			StartMessages: []string{
				"      ✓ Starting on grpc://localhost:9092",
				"      ✓ 3 services registered",
				"      ✓ Protocol buffers loaded",
				"      Status: Serving",
			},
			StatusLabel: "✓ Online",
			Load:        "5 req/min",
			Uptime:      "2h 15m",
			Indicator:   "gRPC Internal",
			StopLabel:   "gRPC service stopped",
		},
		{
			Key:     "ws",
			Name:    "WebSocket Chat Server",
			Address: "ws://localhost:9093",
			StartMessages: []string{
				"      ✓ Starting on ws://localhost:9093",
				"      ✓ Chat rooms initialized",
				"      ✓ User registry ready",
				"      Status: Ready for connections",
			},
			StatusLabel: "✓ Online",
			Load:        "12 users",
			Uptime:      "2h 15m",
			Indicator:   "WebSocket Chat",
			StopLabel:   "WebSocket Chat stopped",
		},
	}
}

// Components returns the current component definitions.
func Components() []Component {
	stateMu.Lock()
	defer stateMu.Unlock()
	copy := make([]Component, len(state.components))
	copy = append(copy[:0], state.components...)
	return copy
}

// MarkRunning records the running components and mode.
func MarkRunning(components []Component) {
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
