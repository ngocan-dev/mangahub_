package output

import (
	"strings"
	"unicode/utf8"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	serviceColumnWidth = 21
	statusColumnWidth  = 10
	addressColumnWidth = 21
	uptimeColumnWidth  = 12
	loadColumnWidth    = 14
	errorColumnWidth   = 28
)

func PrintServerStatusTable(cmd *cobra.Command, st *api.ServerStatus) {
	if st == nil || config.Runtime().Quiet {
		return
	}

	widths := []int{serviceColumnWidth, statusColumnWidth, addressColumnWidth, uptimeColumnWidth, loadColumnWidth, errorColumnWidth}
	headers := []string{"Service", "Status", "Address", "Uptime", "Load", "Error"}

	for i, header := range headers {
		if runes := utf8.RuneCountInString(header) + 1; runes > widths[i] {
			widths[i] = runes
		}
	}

	for _, svc := range st.Services {
		cols := []string{svc.Name, FormatStatus(svc.Status), svc.Address, svc.Uptime, svc.Load, svc.Error}
		for i, col := range cols {
			if runes := utf8.RuneCountInString(col) + 1; runes > widths[i] {
				widths[i] = runes
			}
		}
	}

	cmd.Println("MangaHub Server Status")
	cmd.Println()
	cmd.Println(buildBorder("┌", "┬", "┐", widths))
	cmd.Println(formatRow(headers, widths))
	cmd.Println(buildBorder("├", "┼", "┤", widths))
	for _, svc := range st.Services {
		cmd.Println(formatRow([]string{svc.Name, FormatStatus(svc.Status), svc.Address, svc.Uptime, svc.Load, svc.Error}, widths))
	}
	cmd.Println(buildBorder("└", "┴", "┘", widths))
	cmd.Println()

	cmd.Printf("Overall System Health: %s\n", FormatOverall(st.Overall))
	cmd.Println()

	if len(st.Issues) > 0 {
		cmd.Println("Issues Detected:")
		for _, issue := range st.Issues {
			cmd.Printf("  ✗ %s\n", issue)
		}
		cmd.Println()
	}

	cmd.Println("Database:")
	if st.Database.Connection != "" {
		cmd.Printf("Connection: %s\n", FormatConnection(st.Database.Connection))
	}
	if st.Database.Size != "" {
		cmd.Printf("Size: %s\n", st.Database.Size)
	}
	if len(st.Database.Tables) > 0 {
		cmd.Printf("Tables: %d (%s)\n", len(st.Database.Tables), strings.Join(st.Database.Tables, ", "))
	}
	if st.Database.LastBackup != "" {
		cmd.Printf("Last backup: %s\n", st.Database.LastBackup)
	}
	cmd.Println()

	if st.Resources.Memory != "" {
		cmd.Printf("Memory Usage: %s\n", st.Resources.Memory)
	}
	if st.Resources.CPU != "" {
		cmd.Printf("CPU Usage: %s\n", st.Resources.CPU)
	}
	if st.Resources.Disk != "" {
		cmd.Printf("Disk Space: %s\n", st.Resources.Disk)
	}
}

func buildBorder(start, middle, end string, widths []int) string {
	var b strings.Builder
	b.WriteString(start)
	for i, w := range widths {
		b.WriteString(strings.Repeat("─", w))
		if i == len(widths)-1 {
			b.WriteString(end)
		} else {
			b.WriteString(middle)
		}
	}
	return b.String()
}

func formatRow(cols []string, widths []int) string {
	var b strings.Builder
	b.WriteString("│")
	for i, col := range cols {
		colWidth := widths[i] - 1
		b.WriteString(" ")
		b.WriteString(padRight(col, colWidth))
		b.WriteString("│")
	}
	return b.String()
}

func padRight(s string, width int) string {
	pad := width - utf8.RuneCountInString(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

func FormatStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "online":
		return "✓ Online"
	case "error", "offline", "down":
		return "✗ Error"
	case "warn", "warning":
		return "⚠ Warn"
	default:
		return status
	}
}

func FormatOverall(overall string) string {
	switch strings.ToLower(strings.TrimSpace(overall)) {
	case "healthy", "ok":
		return "✓ Healthy"
	case "degraded":
		return "⚠ Degraded"
	case "offline", "down", "error":
		return "✗ Offline"
	default:
		return overall
	}
}

func FormatConnection(connection string) string {
	switch strings.ToLower(strings.TrimSpace(connection)) {
	case "active", "connected", "online":
		return "✓ Active"
	case "error", "offline", "down":
		return "✗ Error"
	default:
		return connection
	}
}
