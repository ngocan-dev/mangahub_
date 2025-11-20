package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc "google.golang.org/grpc"

	_ "modernc.org/sqlite"

	grpcserver "github.com/ngocan-dev/mangahub/manga-backend/internal/grpc"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/queue"
	"github.com/ngocan-dev/mangahub/manga-backend/internal/tcp"
	pb "github.com/ngocan-dev/mangahub/manga-backend/proto/manga"
)

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
	// Parse command line flags
	address := flag.String("address", ":50051", "gRPC server address")
	dbPath := flag.String("db", "file:data/mangahub.db?_foreign_keys=on", "Database connection string")
	flag.Parse()

	// Open database connection
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize TCP broadcaster for real-time sync
	tcpAddress := os.Getenv("TCP_SERVER_ADDR")
	if tcpAddress == "" {
		tcpAddress = ":9000"
	}
	tcpServer := tcp.NewServer(tcpAddress, 200, db)
	tcpCtx, tcpCancel := context.WithCancel(context.Background())
	defer tcpCancel()
	go startTCPServerWithRestart(tcpCtx, tcpServer, tcpAddress, 200, 5*time.Second)

	writeQueue := queue.NewWriteQueue(1000, 3, nil)
	broadcaster := tcp.NewServerBroadcaster(tcpServer, writeQueue)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register manga service
	mangaServer := grpcserver.NewServer(db)
	mangaServer.SetBroadcaster(broadcaster)
	pb.RegisterMangaServiceServer(grpcServer, mangaServer)

	// Create listener
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("gRPC server listening on %s", *address)
		if err := grpcServer.Serve(lis); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down...", sig)
		grpcServer.GracefulStop()
	case err := <-serverErrChan:
		log.Fatalf("Server error: %v", err)
	}

	log.Println("gRPC server stopped")
}
