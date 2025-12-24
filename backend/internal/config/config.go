package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config holds all environment-driven application settings.
type Config struct {
	DBDriver      string `env:"DB_DRIVER" envRequired:"true"`
	DBDSN         string `env:"DB_DSN" envRequired:"true"`
	MigrationsDir string `env:"MIGRATIONS_DIR" envRequired:"true"`

	DatabaseURL string `env:"DATABASE_URL"`

	RedisURL      string `env:"REDIS_URL"`
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	GRPCServerAddr string `env:"GRPC_SERVER_ADDR"`
	TCPServerAddr  string `env:"TCP_SERVER_ADDR"`
	UDPServerAddr  string `env:"UDP_SERVER_ADDR"`
	WSServerAddr   string `env:"WS_SERVER_ADDR"`
}

// Load loads configuration from a .env file (when present) and the process
// environment, returning a validated Config instance. It fails fast when
// required variables are missing or cannot be parsed.
func Load() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	return &cfg, nil
}

func loadDotEnv() error {
	err := godotenv.Load()
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		// Allow running when .env is not present (e.g. production)
		return nil
	}

	return fmt.Errorf("load .env: %w", err)
}
