package profile

import (
	"os"
	"path/filepath"
	"strings"
)

func humanizePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(path)
	}
	return strings.Replace(filepath.Clean(path), home, "~", 1)
}
