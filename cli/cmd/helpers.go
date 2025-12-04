package cmd

import "github.com/ngocan-dev/mangahub_/cli/config"

// getStoredToken lấy token đã lưu
func getStoredToken() string {
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	return cfg.Token
}

// getStoredUserID lấy user ID đã lưu
func getStoredUserID() int64 {
	cfg, err := config.Load()
	if err != nil {
		return 0
	}
	return cfg.UserID
}

// getStoredConfig lấy toàn bộ config
func getStoredConfig() (*config.Config, error) {
	return config.Load()
}
