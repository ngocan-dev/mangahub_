package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DefaultBaseURL is the fallback API endpoint when none is configured.
const DefaultBaseURL = "http://localhost:8080"

// DefaultGRPCAddress is the default address for gRPC connections.
const DefaultGRPCAddress = "localhost:50051"

// DefaultUDPPort is the default port used for UDP notification delivery.
const DefaultUDPPort = 5050

// Config holds user-specific settings persisted to disk.
type Config struct {
	Token       string   `json:"token"`
	ExpiresAt   string   `json:"expires_at"`
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Settings    struct {
		Autosync      bool `json:"autosync"`
		Notifications bool `json:"notifications"`
	} `json:"settings"`
	BaseURL     string `json:"base_url"`
	GRPCAddress string `json:"grpc_address"`
	UDPPort     int    `json:"udp_port"`
}

// RuntimeOptions capture flags that should not be persisted to disk.
type RuntimeOptions struct {
	Verbose bool
	Quiet   bool
}

// Manager manages loading and saving configuration files.
type Manager struct {
	Path string
	Data Config
}

var (
	currentManager *Manager
	runtime        RuntimeOptions
	mu             sync.RWMutex
)

// SetRuntimeOptions saves the current runtime flags for use across commands.
func SetRuntimeOptions(verbose, quiet bool) {
	mu.Lock()
	defer mu.Unlock()
	runtime = RuntimeOptions{Verbose: verbose, Quiet: quiet}
}

// Runtime returns the current runtime flag configuration.
func Runtime() RuntimeOptions {
	mu.RLock()
	defer mu.RUnlock()
	return runtime
}

// DefaultPath returns the default configuration file path (~/.mangahub/config.json).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mangahub", "config.json"), nil
}

func resolvePath(custom string) (string, error) {
	if custom != "" {
		return custom, nil
	}
	return DefaultPath()
}

// Load loads configuration from disk, creating it with defaults if needed.
func Load(customPath string) (*Manager, error) {
	path, err := resolvePath(customPath)
	if err != nil {
		return nil, err
	}

	if err := ensurePathExists(path); err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := Config{BaseURL: DefaultBaseURL}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = DefaultGRPCAddress
	}

	if cfg.UDPPort == 0 {
		cfg.UDPPort = DefaultUDPPort
	}

	m := &Manager{Path: path, Data: cfg}
	mu.Lock()
	currentManager = m
	mu.Unlock()
	return m, nil
}

// ManagerInstance returns the last loaded manager.
func ManagerInstance() *Manager {
	mu.RLock()
	defer mu.RUnlock()
	return currentManager
}

// Save writes the current configuration back to disk.
func (m *Manager) Save() error {
	if m == nil {
		return errors.New("config manager is not initialized")
	}
	if err := ensurePathExists(m.Path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(m.Path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// UpdateToken saves a new authentication token to the configuration file.
func (m *Manager) UpdateToken(token string) error {
	m.Data.Token = token
	return m.Save()
}

// UpdateSession saves a full authentication session to the configuration file.
func (m *Manager) UpdateSession(token, expiresAt, username string, permissions []string, autosync, notifications bool) error {
	m.Data.Token = token
	m.Data.ExpiresAt = expiresAt
	m.Data.Username = username
	m.Data.Permissions = permissions
	m.Data.Settings.Autosync = autosync
	m.Data.Settings.Notifications = notifications
	return m.Save()
}

// ClearSession removes authentication-related data from the configuration.
func (m *Manager) ClearSession() error {
	m.Data.Token = ""
	m.Data.ExpiresAt = ""
	m.Data.Username = ""
	m.Data.Permissions = nil
	m.Data.Settings.Autosync = false
	m.Data.Settings.Notifications = false

	return m.Save()
}

func ensurePathExists(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		empty := Config{BaseURL: DefaultBaseURL, GRPCAddress: DefaultGRPCAddress, UDPPort: DefaultUDPPort}
		data, err := json.MarshalIndent(empty, "", "  ")
		if err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
