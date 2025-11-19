package db

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// QueryMonitor provides query execution time monitoring
// Search queries complete within acceptable time limits
type QueryMonitor struct {
	slowQueryThreshold time.Duration
	enabled            bool
}

// NewQueryMonitor creates a new query monitor
func NewQueryMonitor(slowQueryThreshold time.Duration) *QueryMonitor {
	return &QueryMonitor{
		slowQueryThreshold: slowQueryThreshold,
		enabled:            true,
	}
}

// MonitorQuery monitors query execution time and logs slow queries
func (qm *QueryMonitor) MonitorQuery(ctx context.Context, queryName string, queryFunc func(context.Context) error) error {
	if !qm.enabled {
		return queryFunc(ctx)
	}

	start := time.Now()
	err := queryFunc(ctx)
	duration := time.Since(start)

	if duration > qm.slowQueryThreshold {
		log.Printf("SLOW QUERY: %s took %v (threshold: %v)", queryName, duration, qm.slowQueryThreshold)
	}

	return err
}

// QueryRowContext wraps sql.DB.QueryRowContext with monitoring
func (qm *QueryMonitor) QueryRowContext(db *sql.DB, ctx context.Context, queryName string, query string, args ...interface{}) *sql.Row {
	if !qm.enabled {
		return db.QueryRowContext(ctx, query, args...)
	}

	start := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > qm.slowQueryThreshold {
		log.Printf("SLOW QUERY: %s took %v (threshold: %v)", queryName, duration, qm.slowQueryThreshold)
	}

	return row
}

// QueryContext wraps sql.DB.QueryContext with monitoring
func (qm *QueryMonitor) QueryContext(db *sql.DB, ctx context.Context, queryName string, query string, args ...interface{}) (*sql.Rows, error) {
	if !qm.enabled {
		return db.QueryContext(ctx, query, args...)
	}

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > qm.slowQueryThreshold {
		log.Printf("SLOW QUERY: %s took %v (threshold: %v)", queryName, duration, qm.slowQueryThreshold)
	}

	return rows, err
}

// ExecContext wraps sql.DB.ExecContext with monitoring
func (qm *QueryMonitor) ExecContext(db *sql.DB, ctx context.Context, queryName string, query string, args ...interface{}) (sql.Result, error) {
	if !qm.enabled {
		return db.ExecContext(ctx, query, args...)
	}

	start := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > qm.slowQueryThreshold {
		log.Printf("SLOW QUERY: %s took %v (threshold: %v)", queryName, duration, qm.slowQueryThreshold)
	}

	return result, err
}
