package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Load reads all configuration from environment variables (optionally via .env)
// and returns a fully parsed Config.
func Load() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	dbDriver, err := getString("DB_DRIVER", "", true)
	if err != nil {
		return nil, err
	}
	dbDSN, err := getString("DB_DSN", "", true)
	if err != nil {
		return nil, err
	}
	migrationsDir, err := getString("MIGRATIONS_DIR", "", true)
	if err != nil {
		return nil, err
	}
	databaseURL, err := getString("DATABASE_URL", "", false)
	if err != nil {
		return nil, err
	}

	redisURL, err := getString("REDIS_URL", "", false)
	if err != nil {
		return nil, err
	}
	redisAddr, err := getString("REDIS_ADDR", "localhost:6379", false)
	if err != nil {
		return nil, err
	}
	redisPassword, err := getString("REDIS_PASSWORD", "", false)
	if err != nil {
		return nil, err
	}
	redisDB, err := getInt("REDIS_DB", 0, false)
	if err != nil {
		return nil, err
	}

	grpcAddr, err := getString("GRPC_SERVER_ADDR", "", false)
	if err != nil {
		return nil, err
	}
	tcpAddr, err := getString("TCP_SERVER_ADDR", "", false)
	if err != nil {
		return nil, err
	}
	udpAddr, err := getString("UDP_SERVER_ADDR", "", false)
	if err != nil {
		return nil, err
	}
	wsAddr, err := getString("WS_SERVER_ADDR", "", false)
	if err != nil {
		return nil, err
	}

	udpMaxClients, udpMaxClientsSet, err := getOptionalInt("UDP_MAX_CLIENTS")
	if err != nil {
		return nil, err
	}
	if !udpMaxClientsSet {
		udpMaxClients = 1000
	}
	udpDisabled := isEnvSet("UDP_SERVER_DISABLED")

	jwtSecret, err := getString("JWT_SECRET", "mangahub-secret-key-change-in-production", false)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		App: AppConfig{
			RedisURL:      redisURL,
			RedisAddr:     redisAddr,
			RedisPassword: redisPassword,
			RedisDB:       redisDB,
			TCPServerAddr: tcpAddr,
			WSServerAddr:  wsAddr,
		},
		DB: DBConfig{
			Driver:        dbDriver,
			DSN:           dbDSN,
			MigrationsDir: migrationsDir,
			DatabaseURL:   databaseURL,
		},
		GRPC: GRPCConfig{
			ServerAddr: grpcAddr,
		},
		UDP: UDPConfig{
			ServerAddr:        udpAddr,
			MaxClients:        udpMaxClients,
			MaxClientsFromEnv: udpMaxClientsSet,
			Disabled:          udpDisabled,
		},
		Auth: AuthConfig{
			JWTSecret: jwtSecret,
		},
	}

	return &cfg, nil
}

func loadDotEnv() error {
	err := loadEnvFiles()
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return fmt.Errorf("load .env: %w", err)
}

func loadEnvFiles(filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	for _, name := range filenames {
		if err := loadEnvFile(name); err != nil {
			return err
		}
	}

	return nil
}

func loadEnvFile(name string) error {
	path := name
	if !filepath.IsAbs(name) {
		if cwd, err := os.Getwd(); err == nil {
			path = filepath.Join(cwd, name)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, strings.Trim(val, `"'`)); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	return scanner.Err()
}

func getString(key, defaultValue string, required bool) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		if required {
			return "", fmt.Errorf("env %s is required", key)
		}
		return defaultValue, nil
	}
	return val, nil
}

func getInt(key string, defaultValue int, required bool) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		if required {
			return 0, fmt.Errorf("env %s is required", key)
		}
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("env %s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func getOptionalInt(key string) (int, bool, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, false, fmt.Errorf("env %s must be an integer: %w", key, err)
	}
	return parsed, true, nil
}

func isEnvSet(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && strings.TrimSpace(val) != ""
}
