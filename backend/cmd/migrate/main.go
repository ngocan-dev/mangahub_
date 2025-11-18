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
	// 1. Kết nối SQLite
	dsn := "file:data/mangahub.db?_foreign_keys=on"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("open db error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db error: %v", err)
	}

	// 2. Đọc folder migration
	migrationsDir := "cmd/db/migrations"

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("read migrations dir error: %v", err)
	}

	// Lấy danh sách file .sql và sort theo tên (001_, 002_, ...)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// 3. Chạy từng file
	for _, name := range files {
		path := filepath.Join(migrationsDir, name)
		log.Printf("Applying migration: %s", path)

		bytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read file error (%s): %v", path, err)
		}

		sqlText := string(bytes)
		// Tách câu lệnh theo dấu ';'
		statements := strings.Split(sqlText, ";")
		for _, stmt := range statements {
			s := strings.TrimSpace(stmt)
			if s == "" {
				continue
			}
			if _, err := db.Exec(s); err != nil {
				log.Fatalf("exec stmt error in %s: %v\nSQL: %s", path, err, s)
			}
		}
	}

	fmt.Println("All migrations applied successfully! DB at data/mangahub.db")
}
