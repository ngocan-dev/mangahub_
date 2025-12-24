//go:build windows
// +build windows

package handlers

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func diskUsage(path string) (string, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(ptr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes); err != nil {
		return "", err
	}

	if totalNumberOfBytes == 0 {
		return "", fmt.Errorf("invalid filesystem size for %s", path)
	}

	used := totalNumberOfBytes - totalNumberOfFreeBytes
	percent := float64(used) / float64(totalNumberOfBytes) * 100

	return fmt.Sprintf("%.1f%% of %s used", percent, formatBytes(int64(totalNumberOfBytes))), nil
}
