package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // Dùng driver giống migration

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/cache"
	dbhealth "github.com/ngocan-dev/mangahub/manga-backend/cmd/db"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/http/handlers"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/middleware"
	"github.com/ngocan-dev/mangahub/manga-backend/cmd/queue"
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

	// Optimize database connection pool for concurrent access
	// API response times remain under 500ms
	// Database queries complete efficiently
	db.SetMaxOpenConns(25)                 // Maximum open connections (SQLite recommends 1, but we use WAL mode)
	db.SetMaxIdleConns(5)                  // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	db.SetConnMaxIdleTime(1 * time.Minute) // Idle connection timeout

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}

	// Enable WAL mode for better concurrent read performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("Warning: Could not enable WAL mode: %v", err)
	}

	// Optimize SQLite for concurrent access
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		log.Printf("Warning: Could not set synchronous mode: %v", err)
	}

	_, err = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Printf("Warning: Could not set busy timeout: %v", err)
	}

	r := gin.Default()

	userHandler := handlers.NewUserHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Initialize Redis cache if available
	// Step 1: System identifies frequently requested manga (handled by cache service)
	var mangaCache *cache.MangaCache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // Default database

	redisClient, err := cache.NewClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Printf("Warning: Redis cache not available: %v. Continuing without cache.", err)
	} else {
		log.Println("Redis cache connected successfully")
		mangaCache = cache.NewMangaCache(redisClient)
		defer redisClient.Close()
	}

	// Initialize database health monitor
	// System attempts automatic reconnection
	healthMonitor := dbhealth.NewHealthMonitor(db, 10*time.Second, 30*time.Second)
	healthMonitor.Start()
	defer healthMonitor.Stop()

	// Initialize write queue for resilience
	// Write operations are queued for later processing
	writeQueue := queue.NewWriteQueue(1000, 3, nil) // Max 1000 operations, 3 retries

	// Create manga service
	mangaService := handlers.GetMangaService(db, mangaCache)

	// Set health monitor and write queue on service
	mangaService.SetDBHealth(healthMonitor)
	mangaService.SetWriteQueue(writeQueue)

	// Initialize write processor
	writeProcessor := queue.NewWriteProcessor(writeQueue, mangaService, db)

	// Set up reconnection callback to process queued operations
	healthMonitor.SetOnReconnect(func() {
		log.Println("Database reconnected, processing queued operations...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		processed, failed := writeProcessor.ProcessAllNow(ctx)
		log.Printf("Processed %d queued operations, %d failed", processed, failed)
	})

	// Start background processor for queued operations
	ctx := context.Background()
	go writeProcessor.StartProcessing(ctx, 30*time.Second)

	mangaHandler := handlers.NewMangaHandlerWithService(db, mangaService)

	// Initialize UDP notifier (optional - can be nil if UDP server not running)
	// In production, you would start the UDP server separately or pass the server instance
	var notificationHandler *handlers.NotificationHandler
	// For now, we'll create it with nil notifier - can be set later if UDP server is available
	notificationHandler = handlers.NewNotificationHandler(db, nil)

	// Initialize rate limiter for handling 50-100 concurrent users
	// API response times remain under 500ms
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute per client

	// Apply rate limiting to all routes
	r.Use(rateLimiter.RateLimitMiddleware())

	// Add request timeout middleware (500ms target)
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Route UC-001: Register
	r.POST("/register", userHandler.Register)

	// Route: Login
	r.POST("/login", authHandler.Login)

	// Route: Search Manga
	r.GET("/manga/search", mangaHandler.Search)

	// Route: Get Manga Details
	r.GET("/manga/:id", mangaHandler.GetDetails)

	// Route: Add Manga to Library
	r.POST("/manga/:id/library", mangaHandler.AddToLibrary)

	// Route: Update Reading Progress
	r.PUT("/manga/:id/progress", mangaHandler.UpdateProgress)

	// Route: Create Review (requires authentication)
	r.POST("/manga/:id/reviews", authHandler.RequireAuth, mangaHandler.CreateReview)

	// Route: Get Reviews
	r.GET("/manga/:id/reviews", mangaHandler.GetReviews)

	// Route: Get Friends Activity Feed (requires authentication)
	r.GET("/friends/activity", authHandler.RequireAuth, mangaHandler.GetFriendsActivityFeed)

	// Route: Get Reading Statistics (requires authentication)
	r.GET("/statistics/reading", authHandler.RequireAuth, mangaHandler.GetReadingStatistics)

	// Route: Get Reading Analytics with filters (requires authentication)
	r.GET("/analytics/reading", authHandler.RequireAuth, mangaHandler.GetReadingAnalytics)

	// Route: Admin - Notify Chapter Release
	r.POST("/admin/notify", notificationHandler.NotifyChapterRelease)

	log.Println("HTTP API listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
