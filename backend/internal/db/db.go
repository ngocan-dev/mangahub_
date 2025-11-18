package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // driver SQLite
)

type Config struct {
	Driver string // "sqlite"
	DSN    string // ví dụ: "file:data/mangahub.db?_foreign_keys=on"
}

func Open(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return db, nil
}
