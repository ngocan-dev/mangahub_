package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	_ "modernc.org/sqlite"

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/internal/config"
	"github.com/ngocan-dev/mangahub/backend/internal/udp"
)

func main() {
	// Parse command line flags
	address := flag.String("address", ":9091", "UDP server address")
	dbPath := flag.String("db", "file:data/mangahub.db?_foreign_keys=on", "Database connection string")
	maxClients := flag.Int("max-clients", 1000, "Maximum concurrent UDP notification clients (0 for unlimited)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Open database connection
	db, err := dbpkg.OpenSQLite(*dbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create UDP server
	server := udp.NewServer(*address, db)
	server.SetMaxClients(*maxClients)

	if cfg.UDP.MaxClientsFromEnv {
		server.SetMaxClients(cfg.UDP.MaxClients)
	}

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
