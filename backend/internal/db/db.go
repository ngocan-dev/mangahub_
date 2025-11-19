package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb" // SQL Server
	_ "modernc.org/sqlite"              // SQLite
)

type Config struct {
	Driver string // "sqlite" hoặc "sqlserver"
	DSN    string // chuỗi kết nối
}

func Open(cfg Config) (*sql.DB, error) {
	// 1. Open database
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db error: %w", err)
	}

	// 2. Tối ưu kết nối theo driver
	switch cfg.Driver {

	case "sqlite":
		// SQLite chỉ trong file → không cần nhiều connections
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

	case "sqlserver":
		// SQL Server chạy tốt với pool lớn hơn
		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(30 * time.Minute)

	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	// 3. Kiểm tra kết nối
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return db, nil
}
