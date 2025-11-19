package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	_ "modernc.org/sqlite"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/grpc"
	pb "github.com/ngocan-dev/mangahub/manga-backend/proto/manga"
)

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

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register manga service
	mangaServer := grpc.NewServer(db)
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
