package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Cấu hình SQLite
	dsn := "file:data/mangahub.db?_foreign_keys=on"

	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	// verify the connection
	if err := database.Ping(); err != nil {
		database.Close()
		log.Fatalf("cannot ping database: %v", err)
	}
	defer database.Close()

	// TODO: truyền *sql.DB này xuống các layer domain/user/manga/...
	// ví dụ:
	// userRepo := user.NewRepository(database)
	// mangaRepo := manga.NewRepository(database)
	// router := http.NewRouter(userRepo, mangaRepo, ...)
	// router.Run(":8080")
}
