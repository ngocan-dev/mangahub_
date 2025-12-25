package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // Dùng driver giống migration

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/domain/friend"
	"github.com/ngocan-dev/mangahub/backend/internal/auth"
	"github.com/ngocan-dev/mangahub/backend/internal/cache"
	"github.com/ngocan-dev/mangahub/backend/internal/config"
	"github.com/ngocan-dev/mangahub/backend/internal/http/handlers"
	"github.com/ngocan-dev/mangahub/backend/internal/middleware"
	"github.com/ngocan-dev/mangahub/backend/internal/queue"
	chapterrepository "github.com/ngocan-dev/mangahub/backend/internal/repository/chapter"
	chapterservice "github.com/ngocan-dev/mangahub/backend/internal/service/chapter"
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
	startTime := time.Now()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	auth.SetSecret(cfg.Auth.JWTSecret)

	db, err := dbpkg.Open(cfg.DB.Driver, cfg.DB.DSN, &dbpkg.PoolConfig{
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
	r.Use(middleware.CORSMiddleware(cfg.App.AllowedOrigins))

	userHandler := handlers.NewUserHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Initialize Redis cache if available
	// Step 1: System identifies frequently requested manga (handled by cache service)
	var mangaCache *cache.MangaCache
	redisAddr := cfg.App.RedisAddr
	redisPassword := cfg.App.RedisPassword
	redisDB := cfg.App.RedisDB

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
	tcpAddress := cfg.App.TCPServerAddr
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

	chapterRepo := chapterrepository.NewRepository(db)
	chapterSvc := chapterservice.NewService(chapterRepo)

	bootstrapCtx, cancelBootstrap := context.WithTimeout(context.Background(), 15*time.Second)
	if err := bootstrapDemoManga(bootstrapCtx, mangaService, chapterSvc); err != nil {
		log.Printf("demo bootstrap failed: %v", err)
	}
	cancelBootstrap()

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
	chapterHandler := handlers.NewChapterHandler(db)

	friendRepo := friend.NewRepository(db)
	friendService := friend.NewService(friendRepo, nil)
	friendHandler := handlers.NewFriendHandler(friendService)

	// Initialize UDP server for chapter release notifications
	udpAddress := cfg.UDP.ServerAddr
	if udpAddress == "" {
		udpAddress = ":9091"
	}
	udpServerEnabled := !cfg.UDP.Disabled
	udpMaxClients := cfg.UDP.MaxClients

	var notificationHandler *handlers.NotificationHandler
	var udpServer *udp.Server
	if udpServerEnabled {
		udpServer = udp.NewServer(udpAddress, db)
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

	wsAddress := cfg.App.WSServerAddr
	if wsAddress == "" {
		wsAddress = ":8081"
	}

	statusHandler := handlers.NewStatusHandler(startTime, db, healthMonitor, writeQueue, cfg.DB.DSN)
	statusHandler.SetTCPServer(tcpServer)
	if udpServer != nil {
		statusHandler.SetUDPServer(udpServer)
	}
	syncHandler := handlers.NewSyncStatusHandler(db, healthMonitor, tcpServer, cfg.DB.DSN)

	apiAddress := ":8080"
	grpcAddress := cfg.GRPC.ServerAddr
	statusHandler.SetAddresses(apiAddress, grpcAddress, tcpAddress, udpAddress)
	statusHandler.SetWSAddress(wsAddress)

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

	r.GET("/server/status", statusHandler.GetStatus)
	r.GET("/sync/status", syncHandler.GetStatus)

	// Route: Login
	r.POST("/login", authHandler.Login)

	// Routes: Friend management
	r.GET("/users/search", authHandler.RequireAuth, friendHandler.Search)
	r.POST("/friends/requests", authHandler.RequireAuth, friendHandler.SendRequest)
	r.POST("/friends/requests/accept", authHandler.RequireAuth, friendHandler.AcceptRequest)

	// Route: Get Popular Manga (cached)
	r.GET("/manga/popular", mangaHandler.GetPopularManga)
	r.GET("/mangas/popular", mangaHandler.GetPopularManga)

	// Route: Search Manga
	r.GET("/manga/search", mangaHandler.Search)

	// Route: Get Manga Details
	r.GET("/manga/:id", mangaHandler.GetDetails)

	// Route: Get Library for authenticated user
	r.GET("/library", authHandler.RequireAuth, mangaHandler.GetLibrary)

	// Route: Get Chapter Details
	r.GET("/chapters/:id", chapterHandler.GetChapter)

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
