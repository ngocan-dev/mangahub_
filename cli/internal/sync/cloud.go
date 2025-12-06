package sync

import "time"

// CloudSyncResult represents the state of a cloud backup layer.
type CloudSyncResult struct {
	Success  bool
	Message  string
	LastSync time.Time
	Pending  int
}

// SyncCloud simulates pushing updates to cloud storage.
func SyncCloud(last time.Time, pending int) CloudSyncResult {
	if pending > 0 {
		return CloudSyncResult{Success: true, Message: "Synced", LastSync: last, Pending: pending}
	}
	return CloudSyncResult{Success: true, Message: "Up to date", LastSync: last}
}
