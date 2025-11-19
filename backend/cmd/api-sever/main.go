package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // Dùng driver giống migration
)

// user handler tối giản
type userHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *userHandler {
	return &userHandler{db: db}
}

func (h *userHandler) Register(c *gin.Context) {
	c.JSON(201, gin.H{"message": "user registered (stub)"})
}

func main() {
	// Mở database từ đúng thư mục Migration
	dsn := "file:data/mangahub.db?_foreign_keys=on"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}

	r := gin.Default()

	userHandler := NewUserHandler(db)

	// Route UC-001: Register
	r.POST("/register", userHandler.Register)

	log.Println("HTTP API listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
