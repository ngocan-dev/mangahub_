package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "modernc.org/sqlite"

	"github.com/ngocan-dev/mangahub/manga-backend/cmd/udp"
)

func main() {
	// Parse command line flags
	address := flag.String("address", ":9091", "UDP server address")
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

	// Create UDP server
	server := udp.NewServer(*address, db)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("Starting UDP notification server on %s", *address)
		if err := server.Start(ctx); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	case err := <-serverErrChan:
		log.Fatalf("Server error: %v", err)
	}

	log.Println("UDP notification server stopped")
}
