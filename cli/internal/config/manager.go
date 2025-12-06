package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RuntimeOptions capture flags that should not be persisted to disk.
type RuntimeOptions struct {
	Verbose bool
	Quiet   bool
}

// Meta stores CLI-wide metadata such as the active profile.
type Meta struct {
	ActiveProfile string `json:"active_profile"`
}

// Manager manages loading and saving configuration files.
type Manager struct {
	Path     string
	MetaPath string
	Data     Config
	meta     Meta
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
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// ConfigDir returns the configuration directory for MangaHub (~/.mangahub).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mangahub"), nil
}

func profileDir(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles", name), nil
}

func profileConfigPath(name string) (string, error) {
	dir, err := profileDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func resolvePath(custom string) (string, error) {
	if custom != "" {
		return custom, nil
	}
	return DefaultPath()
}

func metaPathForConfig(configPath string) (string, error) {
	dir := filepath.Dir(configPath)
	return filepath.Join(dir, "config.meta"), nil
}

// Load loads configuration from disk, creating it with defaults if needed.
func Load(customPath string) (*Manager, error) {
	path, err := resolvePath(customPath)
	if err != nil {
		return nil, err
	}

	metaPath, err := metaPathForConfig(path)
	if err != nil {
		return nil, err
	}

	meta, err := loadMeta(metaPath)
	if err != nil {
		return nil, err
	}

	configPath := path
	active := strings.TrimSpace(meta.ActiveProfile)
	if active == "" {
		active = "default"
	}

	if active != "default" && (customPath == "" || customPath == path) {
		if configPath, err = profileConfigPath(active); err != nil {
			return nil, err
		}
	}

	if err := ensurePathExists(configPath, active); err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig(active)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	if cfg.Auth.Profile == "" {
		cfg.Auth.Profile = active
	}

	applyDerived(&cfg)

	m := &Manager{Path: configPath, MetaPath: metaPath, Data: cfg, meta: meta}
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
	if err := ensurePathExists(m.Path, m.meta.ActiveProfile); err != nil {
		return err
	}
	applyDerived(&m.Data)
	data, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(m.Path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Reset restores the configuration to default values for the active profile.
func (m *Manager) Reset() error {
	active := m.ActiveProfile()
	m.Data = DefaultConfig(active)
	return m.Save()
}

// ActiveProfile returns the currently active profile name.
func (m *Manager) ActiveProfile() string {
	if m == nil {
		return ""
	}
	if strings.TrimSpace(m.meta.ActiveProfile) == "" {
		return "default"
	}
	return m.meta.ActiveProfile
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
	m.Data.Auth.Username = username
	m.Data.Permissions = permissions
	m.Data.Settings.Autosync = autosync
	m.Data.Settings.Notifications = notifications
	m.Data.Notifications.Enabled = notifications
	return m.Save()
}

// ClearSession removes authentication-related data from the configuration.
func (m *Manager) ClearSession() error {
	m.Data.Token = ""
	m.Data.ExpiresAt = ""
	m.Data.Auth.Username = ""
	m.Data.Permissions = nil
	m.Data.Settings.Autosync = false
	m.Data.Settings.Notifications = false
	m.Data.Notifications.Enabled = false

	return m.Save()
}

func ensurePathExists(path, profile string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig(profile)
		data, err := json.MarshalIndent(cfg, "", "  ")
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

func loadMeta(path string) (Meta, error) {
	meta := Meta{ActiveProfile: "default"}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return meta, err
	}

	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := saveMeta(path, meta); err != nil {
			return meta, err
		}
		return meta, nil
	}
	if err != nil {
		return meta, err
	}

	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &meta)
	}
	if strings.TrimSpace(meta.ActiveProfile) == "" {
		meta.ActiveProfile = "default"
	}
	return meta, nil
}

func saveMeta(path string, meta Meta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
