package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProfileList returns existing profiles and the active one.
func ProfileList() ([]string, string, error) {
	metaPath, err := defaultMetaPath()
	if err != nil {
		return nil, "", err
	}
	meta, err := loadMeta(metaPath)
	if err != nil {
		return nil, "", err
	}

	dir, err := ConfigDir()
	if err != nil {
		return nil, "", err
	}
	profiles := []string{"default"}
	profileRoot := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profileRoot)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				profiles = append(profiles, e.Name())
			}
		}
	}

	sort.Strings(profiles)
	return profiles, meta.ActiveProfile, nil
}

// CreateProfile initializes a new profile with default configuration.
func CreateProfile(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("profile name is required")
	}
	if strings.EqualFold(name, "default") {
		return "", fmt.Errorf("profile '%s' already exists", name)
	}

	cfgDir, err := profileDir(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(cfgDir); err == nil {
		return cfgDir, fmt.Errorf("profile '%s' already exists", name)
	}

	if err := os.MkdirAll(filepath.Join(cfgDir, "logs"), 0o755); err != nil {
		return cfgDir, err
	}

	cfg := DefaultConfig(name)
	data, err := jsonMarshal(cfg)
	if err != nil {
		return cfgDir, err
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0o644); err != nil {
		return cfgDir, err
	}

	return cfgDir, nil
}

// SwitchProfile updates the active profile in the meta file.
func SwitchProfile(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("profile name is required")
	}

	metaPath, err := defaultMetaPath()
	if err != nil {
		return "", err
	}

	if strings.EqualFold(name, "default") {
		if err := saveMeta(metaPath, Meta{ActiveProfile: "default"}); err != nil {
			return "", err
		}
		return DefaultPath()
	}

	cfgPath, err := profileConfigPath(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("profile '%s' does not exist", name)
		}
		return "", err
	}

	if err := saveMeta(metaPath, Meta{ActiveProfile: name}); err != nil {
		return "", err
	}
	return cfgPath, nil
}

func defaultMetaPath() (string, error) {
	path, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return metaPathForConfig(path)
}

func jsonMarshal(cfg Config) ([]byte, error) {
	applyDerived(&cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}
