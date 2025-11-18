package main

import (
	"log"

	"github.com/yourname/mangahub_/backend/internal/config"
	"github.com/yourname/mangahub_/backend/internal/db"
	"github.com/yourname/mangahub_/backend/internal/http"
)

func main() {
	cfg := config.Load()
	sqlDB := db.MustConnect(cfg.DB)
	defer sqlDB.Close()

	r := http.NewServer(cfg, sqlDB)
	log.Printf("HTTP API listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
