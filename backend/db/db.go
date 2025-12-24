package db

import (
	"database/sql"
	"fmt"
	"time"
)

// PoolConfig defines connection pool limits for the database.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// defaultPoolConfig returns sane defaults tuned for SQLite WAL mode.
func defaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func resolvePoolConfig(cfg *PoolConfig) PoolConfig {
	poolCfg := defaultPoolConfig()
	if cfg == nil {
		return poolCfg
	}

	if cfg.MaxOpenConns > 0 {
		poolCfg.MaxOpenConns = cfg.MaxOpenConns
	}
	if cfg.MaxIdleConns > 0 {
		poolCfg.MaxIdleConns = cfg.MaxIdleConns
	}
	if cfg.ConnMaxLifetime > 0 {
		poolCfg.ConnMaxLifetime = cfg.ConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime > 0 {
		poolCfg.ConnMaxIdleTime = cfg.ConnMaxIdleTime
	}

	return poolCfg
}

// OpenSQLite opens a SQLite database with performance-focused settings.
// It enables WAL mode for concurrent reads, sets a busy timeout, and configures
// connection pooling to prevent resource exhaustion under load.
func OpenSQLite(dsn string, cfg *PoolConfig) (*sql.DB, error) {
	poolCfg := resolvePoolConfig(cfg)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(poolCfg.MaxOpenConns)
	db.SetMaxIdleConns(poolCfg.MaxIdleConns)
	db.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Enable WAL mode and pragmatic performance settings
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		return nil, fmt.Errorf("set synchronous: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	return db, nil
}

// Open opens a database connection using the provided driver and DSN. For
// SQLite drivers it applies the same performance-focused configuration as
// OpenSQLite. Other drivers are opened with the given pool configuration.
func Open(driver, dsn string, cfg *PoolConfig) (*sql.DB, error) {
	switch driver {
	case "sqlite", "sqlite3":
		return OpenSQLite(dsn, cfg)
	default:
		poolCfg := resolvePoolConfig(cfg)

		db, err := sql.Open(driver, dsn)
		if err != nil {
			return nil, fmt.Errorf("open database: %w", err)
		}

		db.SetMaxOpenConns(poolCfg.MaxOpenConns)
		db.SetMaxIdleConns(poolCfg.MaxIdleConns)
		db.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
		db.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("ping database: %w", err)
		}

		return db, nil
	}
}
