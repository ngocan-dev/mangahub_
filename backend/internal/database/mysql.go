package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	mysqlUser     = "root"
	mysqlPassword = "MySql@2025Dev!"
	mysqlHost     = "localhost"
	mysqlPort     = "3306"
	mysqlDBName   = "mangahub"
)

// OpenMySQL establishes a connection to the MySQL database and configures
// the connection pool. Callers are responsible for closing the returned *sql.DB.
func OpenMySQL() (*sql.DB, error) {
	cfg := mysql.Config{
		User:                 mysqlUser,
		Passwd:               mysqlPassword,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", mysqlHost, mysqlPort),
		DBName:               mysqlDBName,
		ParseTime:            true,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	dsn := cfg.FormatDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: ping database: %w", err)
	}

	return db, nil
}
