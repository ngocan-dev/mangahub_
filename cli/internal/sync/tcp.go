package sync

import "fmt"

// TCPSyncResult represents the outcome of a TCP sync broadcast.
type TCPSyncResult struct {
	Devices int
	Error   error
}

// Broadcast simulates broadcasting updates to connected TCP devices.
func Broadcast(devices int) TCPSyncResult {
	if devices <= 0 {
		return TCPSyncResult{Devices: 0, Error: fmt.Errorf("no devices connected")}
	}
	return TCPSyncResult{Devices: devices}
}
