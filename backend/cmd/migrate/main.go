package main

import (
	"log"
	"os"

	_ "modernc.org/sqlite"

	dbpkg "github.com/ngocan-dev/mangahub/backend/db"
	"github.com/ngocan-dev/mangahub/backend/internal/config"
)

func main() {
	// Support optional "up" subcommand while remaining runnable as `go run ./cmd/migrate`.
	if len(os.Args) > 1 && os.Args[1] != "up" {
		log.Fatalf("unknown command %q (only optional \"up\" is supported)", os.Args[1])
	}

	// Loading config ensures the migrations directory comes from a single source of truth,
	// avoiding brittle hard-coded relative paths that depend on the current working dir.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.DB.MigrationsDir == "" {
		log.Fatal("MIGRATIONS_DIR is not configured")
	}

	db, err := dbpkg.Open(cfg.DB.Driver, cfg.DB.DSN, nil)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := dbpkg.RunMigrations(db, cfg.DB.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}
