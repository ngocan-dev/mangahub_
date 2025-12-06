package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigDir returns the MangaHub configuration directory.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mangahub"), nil
}

// ResolveConfigPath determines the configuration file path to use.
func ResolveConfigPath(custom string) (string, error) {
	if custom != "" {
		return custom, nil
	}

	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// ensureDefaultConfig writes a default config if none exists.
func ensureDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(DefaultConfigYAML), 0o644)
}

// LoadConfig loads configuration from disk or initializes defaults.
func LoadConfig(customPath string) error {
	path, err := ResolveConfigPath(customPath)
	if err != nil {
		return err
	}

	if err := ensureDefaultConfig(path); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	Current = cfg
	Path = path
	return nil
}
