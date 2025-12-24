//go:build !windows
// +build !windows

package handlers

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func diskUsage(path string) (string, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return "", err
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	if total <= 0 {
		return "", fmt.Errorf("invalid filesystem size for %s", path)
	}

	free := int64(stat.Bfree) * int64(stat.Bsize)
	used := total - free
	percent := float64(used) / float64(total) * 100

	return fmt.Sprintf("%.1f%% of %s used", percent, formatBytes(total)), nil
}
