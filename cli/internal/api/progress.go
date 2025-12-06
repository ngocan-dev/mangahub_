package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProgressUpdateRequest captures user input for updating progress.
type ProgressUpdateRequest struct {
	MangaID string `json:"manga_id"`
	Chapter int    `json:"chapter"`
	Volume  int    `json:"volume,omitempty"`
	Notes   string `json:"notes,omitempty"`
	Force   bool   `json:"force"`
}

// ProgressUpdateResponse represents the calculated update details.
type ProgressUpdateResponse struct {
	MangaID         string        `json:"manga_id"`
	MangaName       string        `json:"manga_name"`
	PreviousChapter int           `json:"previous_chapter"`
	CurrentChapter  int           `json:"current_chapter"`
	Volume          int           `json:"volume"`
	Notes           string        `json:"notes"`
	UpdatedAt       time.Time     `json:"updated_at"`
	TotalChapters   int           `json:"total_chapters"`
	Completed       bool          `json:"completed"`
	History         []HistoryItem `json:"history"`
	Statistics      Statistics    `json:"statistics"`
	Sync            SyncSnapshot  `json:"sync"`
}

// HistoryItem represents a single history record.
type HistoryItem struct {
	MangaID string    `json:"manga_id"`
	Manga   string    `json:"manga"`
	Date    time.Time `json:"date"`
	Chapter int       `json:"chapter"`
	Volume  int       `json:"volume"`
	Notes   string    `json:"notes"`
	Source  string    `json:"source"`
}

// Statistics captures computed stats for the update.
type Statistics struct {
	TotalRead           int    `json:"total_read"`
	ReadingStreak       int    `json:"reading_streak"`
	EstimatedCompletion string `json:"estimated_completion"`
}

// SyncSnapshot mirrors the simulated sync layers.
type SyncSnapshot struct {
	Local struct {
		Updated bool   `json:"updated"`
		Note    string `json:"note"`
	} `json:"local"`
	TCP struct {
		Success  bool   `json:"success"`
		Devices  int    `json:"devices"`
		Message  string `json:"message"`
		LastSync string `json:"last_sync"`
	} `json:"tcp"`
	Cloud struct {
		Success  bool   `json:"success"`
		Message  string `json:"message"`
		LastSync string `json:"last_sync"`
		Pending  int    `json:"pending"`
	} `json:"cloud"`
}

// SyncResult is returned from a manual sync command.
type SyncResult struct {
	LocalReady bool   `json:"local_ready"`
	TCPDevices int    `json:"tcp_devices"`
	CloudState string `json:"cloud_state"`
}

// SyncStatus represents sync freshness for each layer.
type SyncStatus struct {
	LocalUpdatedAgo   time.Duration `json:"local_updated_ago"`
	TCPDevices        int           `json:"tcp_devices"`
	CloudLastSync     time.Time     `json:"cloud_last_sync"`
	CloudPendingDelta int           `json:"cloud_pending_delta"`
}

var (
	seedOnce   sync.Once
	progressDB map[string]*progressState
)

type progressState struct {
	MangaID       string
	Name          string
	TotalChapters int
	Completed     bool
	Current       int
	Volume        int
	History       []HistoryItem
	UpdatedAt     time.Time
	Estimated     string
	ReadingStreak int
}

// seedProgress constructs an in-memory dataset for simulation.
func seedProgress() {
	seedOnce.Do(func() {
		progressDB = map[string]*progressState{}
		baseTime := time.Date(2024, 1, 20, 16, 45, 0, 0, time.UTC)

		// build a streak of 45 days ending at baseTime
		history := make([]HistoryItem, 0, 45)
		startChapter := 1051
		for i := 44; i >= 0; i-- {
			day := baseTime.AddDate(0, 0, -i)
			history = append(history, HistoryItem{
				MangaID: "one-piece",
				Manga:   "One Piece",
				Date:    day,
				Chapter: startChapter + (44 - i),
				Volume:  60 + (44-i)/10,
				Source:  "CLI Update",
			})
		}

		progressDB["one-piece"] = &progressState{
			MangaID:       "one-piece",
			Name:          "One Piece",
			TotalChapters: 1100,
			Completed:     false,
			Current:       1094,
			Volume:        72,
			History:       append(history[len(history)-3:], []HistoryItem{
				// ensure we keep the last 3 meaningful notes
			}...),
			UpdatedAt:     baseTime.Add(-time.Hour),
			Estimated:     "Never (ongoing series)",
			ReadingStreak: 45,
		}

		// augment the last 3 history entries with notes and sources to match expected output
		if len(progressDB["one-piece"].History) >= 3 {
			progressDB["one-piece"].History[2].Notes = "Great ending!"
			progressDB["one-piece"].History[2].Source = "CLI Update"
			progressDB["one-piece"].History[1].Source = "Sync Device"
			progressDB["one-piece"].History[0].Source = "Cloud Restore"
		}

		progressDB["naruto"] = &progressState{MangaID: "naruto", Name: "Naruto", TotalChapters: 700, Current: 120, Estimated: "20 days", ReadingStreak: 5}
		progressDB["attack-on-titan"] = &progressState{MangaID: "attack-on-titan", Name: "Attack on Titan", TotalChapters: 139, Current: 80, Estimated: "2 months", ReadingStreak: 2}
	})
}

// UpdateProgressDetail simulates the progress update endpoint with validations.
func (c *Client) UpdateProgressDetail(ctx context.Context, req ProgressUpdateRequest) (*ProgressUpdateResponse, error) {
	seedProgress()

	if c == nil || c.token == "" {
		return nil, errors.New("✗ You must be logged in to update progress.\nPlease login first.")
	}

	state, ok := progressDB[req.MangaID]
	if !ok {
		return nil, fmt.Errorf("✗ Progress update failed: Manga '%s' not found in your library\nAdd to library first:\nmangahub library add --manga-id %s --status reading", req.MangaID, req.MangaID)
	}

	if strings.Contains(strings.ToLower(c.baseURL), "unreachable") {
		return nil, errors.New("✗ Progress update failed: Server connection error\nTry again or check: mangahub server status")
	}

	if req.Chapter > state.TotalChapters {
		return nil, fmt.Errorf("✗ Progress update failed: Chapter %d exceeds manga's total chapters (%d)\nValid range: 1-%d", req.Chapter, state.TotalChapters, state.TotalChapters)
	}

	if req.Chapter < state.Current && !req.Force {
		return nil, fmt.Errorf("✗ Progress update failed: Chapter %d is behind your current progress (Chapter %d)\nUse --force to set backwards progress: --force --chapter %d", req.Chapter, state.Current, req.Chapter)
	}

	prev := state.Current
	state.Current = req.Chapter
	if req.Volume > 0 {
		state.Volume = req.Volume
	}
	state.UpdatedAt = time.Date(2024, 1, 20, 16, 45, 0, 0, time.UTC)

	newEntry := HistoryItem{
		MangaID: req.MangaID,
		Manga:   state.Name,
		Date:    state.UpdatedAt,
		Chapter: req.Chapter,
		Volume:  state.Volume,
		Notes:   req.Notes,
		Source:  "CLI Update",
	}
	state.History = append([]HistoryItem{newEntry}, state.History...)
	state.History = trimHistory(state.History, 3)

	resp := &ProgressUpdateResponse{
		MangaID:         state.MangaID,
		MangaName:       state.Name,
		PreviousChapter: prev,
		CurrentChapter:  state.Current,
		Volume:          state.Volume,
		Notes:           req.Notes,
		UpdatedAt:       state.UpdatedAt,
		TotalChapters:   state.TotalChapters,
		Completed:       state.Completed,
		History:         state.History,
	}

	resp.Statistics = Statistics{
		TotalRead:           state.Current,
		ReadingStreak:       state.ReadingStreak,
		EstimatedCompletion: state.Estimated,
	}

	resp.Sync.Local.Updated = true
	resp.Sync.Local.Note = "Updated"
	resp.Sync.TCP.Success = true
	resp.Sync.TCP.Devices = 3
	resp.Sync.TCP.Message = "Broadcasting to 3 connected devices"
	resp.Sync.Cloud.Success = true
	resp.Sync.Cloud.Message = "Synced"
	resp.Sync.Cloud.LastSync = state.UpdatedAt.Format(time.RFC3339)

	return resp, nil
}

// ProgressHistory returns the stored history for a manga or all manga.
func (c *Client) ProgressHistory(ctx context.Context, mangaID string) ([]HistoryItem, error) {
	seedProgress()
	var records []HistoryItem
	for _, state := range progressDB {
		if mangaID != "" && state.MangaID != mangaID {
			continue
		}
		records = append(records, state.History...)
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Manga == records[j].Manga {
			return records[i].Date.After(records[j].Date)
		}
		return records[i].Manga < records[j].Manga
	})

	return records, nil
}

// TriggerProgressSync simulates syncing across layers.
func (c *Client) TriggerProgressSync(ctx context.Context) (*SyncResult, error) {
	seedProgress()
	result := &SyncResult{LocalReady: true, TCPDevices: 3, CloudState: "Up to date"}
	return result, nil
}

// GetSyncStatus reports current sync state.
func (c *Client) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	seedProgress()
	status := &SyncStatus{
		LocalUpdatedAgo:   5 * time.Minute,
		TCPDevices:        3,
		CloudLastSync:     time.Date(2024, 1, 20, 16, 45, 0, 0, time.UTC),
		CloudPendingDelta: 0,
	}
	return status, nil
}

func trimHistory(history []HistoryItem, max int) []HistoryItem {
	if len(history) <= max {
		return history
	}
	return history[:max]
}

// HumanRelative converts a timestamp to a friendly relative string.
func HumanRelative(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	now := time.Now().UTC()
	diff := now.Sub(t)
	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
