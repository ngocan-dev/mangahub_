package udp

import (
	"sync/atomic"
	"time"
)

// Metrics tracks UDP server statistics
type Metrics struct {
	// Counters - tích lũy theo thời gian
	TotalRegistrations   atomic.Int64
	TotalUnregistrations atomic.Int64
	TotalNotifications   atomic.Int64
	FailedNotifications  atomic.Int64
	PacketsReceived      atomic.Int64
	PacketsSent          atomic.Int64

	// Gauges - giá trị hiện tại
	ActiveClients atomic.Int64

	// Metadata
	StartTime time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// GetStats returns current metrics as a map
func (m *Metrics) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"uptime_seconds":         time.Since(m.StartTime).Seconds(),
		"total_registrations":    m.TotalRegistrations.Load(),
		"total_unregistrations":  m.TotalUnregistrations.Load(),
		"total_notifications":    m.TotalNotifications.Load(),
		"failed_notifications":   m.FailedNotifications.Load(),
		"packets_received":       m.PacketsReceived.Load(),
		"packets_sent":           m.PacketsSent.Load(),
		"active_clients":         m.ActiveClients.Load(),
	}
}

// IncrementRegistrations increments registration counter
func (m *Metrics) IncrementRegistrations() {
	m.TotalRegistrations.Add(1)
}

// IncrementUnregistrations increments unregistration counter
func (m *Metrics) IncrementUnregistrations() {
	m.TotalUnregistrations.Add(1)
}

// IncrementNotifications increments notification counter
func (m *Metrics) IncrementNotifications() {
	m.TotalNotifications.Add(1)
}

// IncrementFailedNotifications increments failed notification counter
func (m *Metrics) IncrementFailedNotifications() {
	m.FailedNotifications.Add(1)
}

// IncrementPacketsReceived increments packets received counter
func (m *Metrics) IncrementPacketsReceived() {
	m.PacketsReceived.Add(1)
}

// IncrementPacketsSent increments packets sent counter
func (m *Metrics) IncrementPacketsSent() {
	m.PacketsSent.Add(1)
}

// SetActiveClients sets current active clients count
func (m *Metrics) SetActiveClients(count int64) {
	m.ActiveClients.Store(count)
}

// GetUptime returns server uptime duration
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}
