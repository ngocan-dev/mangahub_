package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	// 1. Create DB directory
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data folder: %v", err)
	}

	// 2. Connect to SQLite
	dsn := "file:data/mangahub.db?_foreign_keys=on"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Open DB error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Ping DB error: %v", err)
	}

	log.Println("Connected to SQLite at data/mangahub.db")

	// 3. Read migrations folder
	migrationsDir := "cmd/db/migrations"

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Read migrations dir error: %v", err)
	}

	// Find all *.sql files
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}

	if len(files) == 0 {
		log.Fatal("No migration files found")
	}

	// Sort ascending: 001_, 002_, ...
	sort.Strings(files)

	// 4. Apply each migration inside its own transaction
	for _, filename := range files {

		path := filepath.Join(migrationsDir, filename)
		log.Printf("⬆️ Applying migration: %s", filename)

		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Read file error %s: %v", filename, err)
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Begin txn error: %v", err)
		}

		// Exec entire file (SQLite allows multiple statements)
		_, err = tx.Exec(string(sqlBytes))
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf(
				"Migration failed (%s): %v\n--- SQL CONTENT ---\n%s",
				filename, err, string(sqlBytes),
			)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Commit error: %v", err)
		}

		log.Printf("Migration applied: %s\n", filename)
	}

	fmt.Println("All migrations applied successfully!")
}
