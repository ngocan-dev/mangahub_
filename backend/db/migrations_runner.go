package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

// RunMigrations applies all SQL migrations in the provided directory.
// Migrations are applied in lexicographic order and recorded in the
// schema_migrations table to make the process idempotent.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if migrationsDir == "" {
		return fmt.Errorf("migrations directory is required")
	}

	if _, err := db.Exec(schemaMigrationsTable); err != nil {
		return fmt.Errorf("prepare schema_migrations table: %w", err)
	}

	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}

	sort.Strings(files)

	for _, file := range files {
		if _, ok := applied[file]; ok {
			continue
		}

		path := filepath.Join(migrationsDir, file)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", file, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", file, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}
	}

	return nil
}

func loadAppliedMigrations(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}
