package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Minimal local handler implementation to avoid external package dependency.
// In production you should move this to a dedicated package and implement real logic.
type userHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *userHandler {
	return &userHandler{db: db}
}

func (h *userHandler) Register(c *gin.Context) {
	// Placeholder implementation: respond with 201 Created
	c.JSON(201, gin.H{"message": "user registered (stub)"})
}

func main() {
	// open the sqlite database (file name can be changed as needed)
	db, err := sql.Open("sqlite3", "mangahub.db")
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}

	r := gin.Default()

	userHandler := NewUserHandler(db)
	// route UC-001
	r.POST("/register", userHandler.Register)

	log.Println("HTTP API listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
