package sync

import "time"

// LayerStatus represents a local sync layer state.
type LayerStatus struct {
	OK        bool
	Message   string
	Timestamp time.Time
}

// ManualSyncSummary aggregates the results of a manual sync run.
type ManualSyncSummary struct {
	Local LayerStatus
	TCP   TCPSyncResult
	Cloud CloudSyncResult
}

// RunManualSync performs simulated sync actions across layers.
func RunManualSync(devices int, lastSync time.Time, pending int) ManualSyncSummary {
	now := time.Now().UTC()
	tcpResult := Broadcast(devices)
	cloud := SyncCloud(lastSync, pending)

	return ManualSyncSummary{
		Local: LayerStatus{OK: true, Message: "Updated", Timestamp: now},
		TCP:   tcpResult,
		Cloud: cloud,
	}
}
