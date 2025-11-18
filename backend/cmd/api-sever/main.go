package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ngocan-dev/mangahub_/backend/cmd/http/handlers"
)

func main() {
	// Dùng lại DSN SQLite bạn đã dùng cho migrate
	dsn := "file:data/mangahub.db?_foreign_keys=on"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}

	r := gin.Default()

	userHandler := handlers.NewUserHandler(db)
	// route UC-001
	r.POST("/register", userHandler.Register)

	log.Println("HTTP API listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
