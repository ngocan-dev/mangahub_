package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/domain/friend"
	"github.com/ngocan-dev/mangahub/backend/domain/user"
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
	ws "github.com/ngocan-dev/mangahub/backend/internal/websocket"
)

func startTCPServerWithRestart(
	ctx context.Context,
	server *tcp.Server,
	address string,
	maxClients int,
	backoff time.Duration,
) {
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

	// Root context for the whole process (graceful shutdown)
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Println("DB DRIVER =", cfg.DB.Driver)
	log.Println("DB DSN =", cfg.DB.DSN)

	// Auth secret
	auth.SetSecret(cfg.Auth.JWTSecret)

	// DB
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

	// Gin
	r := gin.Default()
	r.Use(middleware.CORSMiddleware(cfg.App.AllowedOrigins))

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(rateLimiter.RateLimitMiddleware())

	// Request timeout (adjust if too aggressive for your DB queries)
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Handlers
	userHandler := handlers.NewUserHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Optional Redis cache
	var mangaCache *cache.MangaCache
	redisClient, err := cache.NewClient(cfg.App.RedisAddr, cfg.App.RedisPassword, cfg.App.RedisDB)
	if err != nil {
		log.Printf("Warning: Redis cache not available: %v. Continuing without cache.", err)
	} else {
		log.Println("Redis cache connected successfully")
		mangaCache = cache.NewMangaCache(redisClient)
		defer redisClient.Close()
	}

	// DB health monitor
	healthMonitor := dbpkg.NewHealthMonitor(db, 10*time.Second, 30*time.Second)
	healthMonitor.Start()
	defer healthMonitor.Stop()

	// Write queue
	writeQueue := queue.NewWriteQueue(1000, 3, nil)

	// TCP progress sync server
	tcpAddress := cfg.App.TCPServerAddr
	if tcpAddress == "" {
		tcpAddress = ":9000"
	}
	tcpServer := tcp.NewServer(tcpAddress, 200, db)

	go startTCPServerWithRestart(rootCtx, tcpServer, tcpAddress, 200, 5*time.Second)

	broadcaster := tcp.NewServerBroadcaster(tcpServer, writeQueue)

	// Manga service + handlers
	mangaService := handlers.GetMangaService(db, mangaCache)
	mangaService.SetDBHealth(healthMonitor)
	mangaService.SetWriteQueue(writeQueue)

	chapterRepo := chapterrepository.NewRepository(db)
	chapterSvc := chapterservice.NewService(chapterRepo)

	// Demo bootstrap (if enabled)
	if cfg.EnableDemoData {
		log.Println("⚠️ Loading DEMO data from SQLite")
		bootstrapCtx, cancel := context.WithTimeout(rootCtx, 15*time.Second)
		if err := bootstrapDemoManga(bootstrapCtx, mangaService, chapterSvc); err != nil {
			log.Printf("demo bootstrap failed: %v", err)
		}
		cancel()
	} else {
		log.Println("✅ Demo bootstrap skipped")
	}

	// Write processor
	writeProcessor := queue.NewWriteProcessor(writeQueue, mangaService, db, broadcaster)

	healthMonitor.SetOnReconnect(func() {
		log.Println("Database reconnected, processing queued operations...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		processed, failed := writeProcessor.ProcessAllNow(ctx)
		log.Printf("Processed %d queued operations, %d failed", processed, failed)
	})

	go writeProcessor.StartProcessing(rootCtx, 30*time.Second)

	mangaHandler := handlers.NewMangaHandlerWithService(db, mangaService)
	mangaHandler.SetBroadcaster(broadcaster)
	mangaHandler.SetDBHealth(healthMonitor)
	mangaHandler.SetWriteQueue(writeQueue)

	chapterHandler := handlers.NewChapterHandler(db)

	// Friend domain wiring
	userRepo := user.NewRepository(db)
	friendRepo := friend.NewRepository(db)
	friendService := friend.NewService(friendRepo, userRepo, nil) // consider a Noop notifier instead of nil
	friendHandler := handlers.NewFriendHandler(friendService)

	// WebSocket chat
	chatHub := ws.NewDirectChatHub(db)
	chatHandler := handlers.NewChatHandler(chatHub)

	// UDP notification server
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

		go func() {
			log.Printf("Starting UDP notification server on %s", udpAddress)
			if err := udpServer.Start(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
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

	// Status/sync
	apiAddress := ":8080"
	grpcAddress := cfg.GRPC.ServerAddr

	statusHandler := handlers.NewStatusHandler(startTime, db, healthMonitor, writeQueue, cfg.DB.DSN)
	statusHandler.SetTCPServer(tcpServer)
	if udpServer != nil {
		statusHandler.SetUDPServer(udpServer)
	}
	statusHandler.SetAddresses(apiAddress, grpcAddress, tcpAddress, udpAddress)
	statusHandler.SetWSAddress(wsAddress)

	syncHandler := handlers.NewSyncStatusHandler(db, healthMonitor, tcpServer, cfg.DB.DSN)

	// --------------------
	// Routes
	// --------------------

	// UC-001: Register
	r.POST("/register", userHandler.Register)

	// Status/sync
	r.GET("/server/status", statusHandler.GetStatus)
	r.GET("/sync/status", syncHandler.GetStatus)

	// Login
	r.POST("/login", authHandler.Login)
	r.GET("/me", authHandler.RequireAuth, authHandler.Me)

	// Friend management
	r.GET("/users/search", authHandler.RequireAuth, friendHandler.Search)
	r.GET("/friends", authHandler.RequireAuth, friendHandler.ListFriends)
	r.GET("/friends/requests", authHandler.RequireAuth, friendHandler.PendingRequests)
	r.POST("/friends/request", authHandler.RequireAuth, friendHandler.SendRequest)
	r.POST("/friends/accept", authHandler.RequireAuth, friendHandler.AcceptRequest)
	r.POST("/friends/reject", authHandler.RequireAuth, friendHandler.RejectRequest)

	// Legacy paths (optional)
	r.POST("/friends/requests", authHandler.RequireAuth, friendHandler.SendRequest)
	r.POST("/friends/requests/accept", authHandler.RequireAuth, friendHandler.AcceptRequest)

	// WebSocket chat
	r.GET("/ws/chat", authHandler.RequireAuth, chatHandler.Serve)

	// Manga
	r.GET("/manga/popular", mangaHandler.GetPopularManga)
	r.GET("/mangas/popular", mangaHandler.GetPopularManga)

	r.GET("/mangas/search", mangaHandler.Search)
	r.GET("/mangas/:id", mangaHandler.GetDetails)

	r.GET("/library", authHandler.RequireAuth, mangaHandler.GetLibrary)
	r.POST("/mangas/:id/library", authHandler.RequireAuth, mangaHandler.AddToLibrary)

	r.GET("/chapters/:id", chapterHandler.GetChapter)

	r.PUT("/mangas/:id/progress", authHandler.RequireAuth, mangaHandler.UpdateProgress)

	r.POST("/mangas/:id/reviews", authHandler.RequireAuth, mangaHandler.CreateReview)
	r.GET("/mangas/:id/reviews", mangaHandler.GetReviews)

	// r.GET("/friends/activity", authHandler.RequireAuth, mangaHandler.GetFriendsActivityFeed)

	r.GET("/statistics/reading", authHandler.RequireAuth, mangaHandler.GetReadingStatistics)
	r.GET("/analytics/reading", authHandler.RequireAuth, mangaHandler.GetReadingAnalytics)

	// Admin notify
	r.POST("/admin/notify", authHandler.RequireAuth, notificationHandler.NotifyChapterRelease)

	// --------------------
	// HTTP server (graceful shutdown)
	// --------------------
	srv := &http.Server{
		Addr:    apiAddress,
		Handler: r,
	}

	go func() {
		log.Println("HTTP API listening on", apiAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-rootCtx.Done()
	log.Println("Shutdown signal received, shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped")
	}
}
