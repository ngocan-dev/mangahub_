package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // Dùng driver giống migration

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/domain/friend"
	"github.com/ngocan-dev/mangahub/backend/internal/cache"
	"github.com/ngocan-dev/mangahub/backend/internal/http/handlers"
	"github.com/ngocan-dev/mangahub/backend/internal/middleware"
	"github.com/ngocan-dev/mangahub/backend/internal/queue"
	"github.com/ngocan-dev/mangahub/backend/internal/tcp"
	"github.com/ngocan-dev/mangahub/backend/internal/udp"
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

func startTCPServerWithRestart(ctx context.Context, server *tcp.Server, address string, maxClients int, backoff time.Duration) {
	for {
		log.Printf("Starting TCP server on %s (max clients: %d)", address, maxClients)
		if err := server.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("TCP server stopped with error: %v", err)
		} else {
			log.Printf("TCP server stopped")
		}

		if ctx.Err() != nil {
			return
		}

		log.Printf("Restarting TCP server in %s...", backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
	}
}

func main() {
	// Mở database từ đúng thư mục Migration
	dsn := "file:data/mangahub.db?_foreign_keys=on"

	db, err := dbpkg.OpenSQLite(dsn, &dbpkg.PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

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
	healthMonitor := dbpkg.NewHealthMonitor(db, 10*time.Second, 30*time.Second)
	healthMonitor.Start()
	defer healthMonitor.Stop()

	// Initialize write queue for resilience
	// Write operations are queued for later processing
	writeQueue := queue.NewWriteQueue(1000, 3, nil) // Max 1000 operations, 3 retries

	// Start TCP broadcaster for progress updates
	tcpAddress := os.Getenv("TCP_SERVER_ADDR")
	if tcpAddress == "" {
		tcpAddress = ":9000"
	}
	tcpServer := tcp.NewServer(tcpAddress, 200, db)
	tcpCtx, tcpCancel := context.WithCancel(context.Background())
	defer tcpCancel()
	go startTCPServerWithRestart(tcpCtx, tcpServer, tcpAddress, 200, 5*time.Second)

	broadcaster := tcp.NewServerBroadcaster(tcpServer, writeQueue)

	// Create manga service
	mangaService := handlers.GetMangaService(db, mangaCache)

	// Set health monitor and write queue on service
	mangaService.SetDBHealth(healthMonitor)
	mangaService.SetWriteQueue(writeQueue)

	// Initialize write processor
	writeProcessor := queue.NewWriteProcessor(writeQueue, mangaService, db, broadcaster)

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
	mangaHandler.SetBroadcaster(broadcaster)
	mangaHandler.SetDBHealth(healthMonitor)
	mangaHandler.SetWriteQueue(writeQueue)

	friendRepo := friend.NewRepository(db)
	friendService := friend.NewService(friendRepo, nil)
	friendHandler := handlers.NewFriendHandler(friendService)

	// Initialize UDP server for chapter release notifications
	udpAddress := os.Getenv("UDP_SERVER_ADDR")
	if udpAddress == "" {
		udpAddress = ":9091"
	}
	udpServerEnabled := os.Getenv("UDP_SERVER_DISABLED") == ""
	udpMaxClients := 1000

	if maxClientsEnv := os.Getenv("UDP_MAX_CLIENTS"); maxClientsEnv != "" {
		if maxFromEnv, err := strconv.Atoi(maxClientsEnv); err == nil {
			udpMaxClients = maxFromEnv
		}
	}

	var notificationHandler *handlers.NotificationHandler
	if udpServerEnabled {
		udpServer := udp.NewServer(udpAddress, db)
		udpServer.SetMaxClients(udpMaxClients)
		udpCtx, udpCancel := context.WithCancel(context.Background())
		defer udpCancel()
		go func() {
			log.Printf("Starting UDP notification server on %s", udpAddress)
			if err := udpServer.Start(udpCtx); err != nil {
				log.Printf("UDP server stopped: %v", err)
			}
		}()

		notifier := udp.NewNotifier(udpServer)
		notificationHandler = handlers.NewNotificationHandler(db, notifier)
	} else {
		log.Println("UDP notification server disabled; chapter notifications will be unavailable")
		notificationHandler = handlers.NewNotificationHandler(db, nil)
	}

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

	// Routes: Friend management
	r.GET("/users/search", authHandler.RequireAuth, friendHandler.Search)
	r.POST("/friends/requests", authHandler.RequireAuth, friendHandler.SendRequest)
	r.POST("/friends/requests/accept", authHandler.RequireAuth, friendHandler.AcceptRequest)

	// Route: Get Popular Manga (cached)
	r.GET("/manga/popular", mangaHandler.GetPopularManga)

	// Route: Search Manga
	r.GET("/manga/search", mangaHandler.Search)

	// Route: Get Manga Details
	r.GET("/manga/:id", mangaHandler.GetDetails)

	// Route: Add Manga to Library
	r.POST("/manga/:id/library", mangaHandler.AddToLibrary)

	// Route: Update Reading Progress (requires authentication)
	r.PUT("/manga/:id/progress", authHandler.RequireAuth, mangaHandler.UpdateProgress)

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

	// Route: Admin - Notify Chapter Release (requires authentication)
	r.POST("/admin/notify", authHandler.RequireAuth, notificationHandler.NotifyChapterRelease)

	log.Println("HTTP API listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
