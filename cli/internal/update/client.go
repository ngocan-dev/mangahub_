package update

// Info describes available update metadata.
type Info struct {
	CurrentVersion string   `json:"current_version"`
	LatestVersion  string   `json:"latest_version"`
	ReleaseNotes   []string `json:"release_notes"`
	UpToDate       bool     `json:"up_to_date"`
}

// InstallResult captures the outcome of an update attempt.
type InstallResult struct {
	FromVersion string   `json:"from_version"`
	ToVersion   string   `json:"to_version"`
	Steps       []string `json:"steps"`
	Success     bool     `json:"success"`
}

const latestVersion = "1.4.0"

// CheckForUpdates simulates contacting a remote endpoint for update metadata.
func CheckForUpdates(current string) Info {
	info := Info{
		CurrentVersion: current,
		LatestVersion:  latestVersion,
		ReleaseNotes: []string{
			"Improved progress syncing",
			"New statistics breakdown",
		},
	}
	info.UpToDate = current == latestVersion
	if info.UpToDate {
		info.ReleaseNotes = []string{}
	}
	return info
}

// InstallLatest simulates downloading and installing the latest CLI version.
func InstallLatest(current string) InstallResult {
	return InstallResult{
		FromVersion: current,
		ToVersion:   latestVersion,
		Steps: []string{
			"Downloaded version " + latestVersion,
			"Verified checksum",
			"Installed new binary",
		},
		Success: true,
	}
}

// LatestVersion returns the latest available CLI version.
func LatestVersion() string {
	return latestVersion
}
