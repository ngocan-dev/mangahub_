package db

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// HealthMonitor monitors database connection health
type HealthMonitor struct {
	db                *sql.DB
	mu                sync.RWMutex
	isHealthy         bool
	lastCheck         time.Time
	checkInterval     time.Duration
	reconnectInterval time.Duration
	stopChan          chan struct{}
	onReconnect       func() // Callback when reconnected
}

// NewHealthMonitor creates a new database health monitor
func NewHealthMonitor(db *sql.DB, checkInterval, reconnectInterval time.Duration) *HealthMonitor {
	hm := &HealthMonitor{
		db:                db,
		isHealthy:         true, // Assume healthy initially
		checkInterval:     checkInterval,
		reconnectInterval: reconnectInterval,
		stopChan:          make(chan struct{}),
	}

	// Initial health check
	hm.checkHealth()

	return hm
}

// IsHealthy returns whether the database is currently healthy
func (hm *HealthMonitor) IsHealthy() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.isHealthy
}

// Start begins monitoring the database health
func (hm *HealthMonitor) Start() {
	go hm.monitor()
}

// Stop stops monitoring
func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
}

// SetOnReconnect sets a callback function to be called when database reconnects
func (hm *HealthMonitor) SetOnReconnect(callback func()) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.onReconnect = callback
}

// checkHealth performs a health check
func (hm *HealthMonitor) checkHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := hm.db.PingContext(ctx)
	wasHealthy := hm.isHealthy

	hm.mu.Lock()
	if err != nil {
		hm.isHealthy = false
	} else {
		// If we were unhealthy and now we're healthy, trigger reconnect callback
		if !wasHealthy && hm.isHealthy == false {
			hm.isHealthy = true
			if hm.onReconnect != nil {
				go hm.onReconnect()
			}
		}
		hm.isHealthy = true
	}
	hm.lastCheck = time.Now()
	hm.mu.Unlock()

	return hm.isHealthy
}

// monitor continuously monitors database health
func (hm *HealthMonitor) monitor() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.checkHealth()
		}
	}
}

// AttemptReconnect attempts to reconnect to the database
func (hm *HealthMonitor) AttemptReconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := hm.db.PingContext(ctx)
	if err == nil {
		hm.mu.Lock()
		wasHealthy := hm.isHealthy
		hm.isHealthy = true
		if !wasHealthy && hm.onReconnect != nil {
			go hm.onReconnect()
		}
		hm.mu.Unlock()
	}

	return err
}
