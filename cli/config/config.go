package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config lưu thông tin cấu hình CLI
type Config struct {
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	APIURL   string `json:"api_url"`
}

// GetConfigPath trả về đường dẫn file config
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".mangahub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load đọc config từ file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File chưa tồn tại, trả về config mặc định
			return &Config{
				APIURL: "http://localhost:8080",
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set mặc định nếu chưa có
	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:8080"
	}

	return &cfg, nil
}

// Save lưu config vào file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// Clear xóa config
func Clear() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
