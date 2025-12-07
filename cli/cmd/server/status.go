package server

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show server status",
	Long:    "Display the current status of the MangaHub server.",
	Example: "mangahub server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		degraded, _ := cmd.Flags().GetBool("degraded")
		status := buildStatusSummary(degraded)
		if format == output.FormatJSON {
			output.PrintJSON(cmd, status)
			return nil
		}
		if config.Runtime().Verbose {
			output.PrintJSON(cmd, status)
			return nil
		}
		if config.Runtime().Quiet {
			return nil
		}
		if degraded {
			printDegradedStatus(cmd)
			return nil
		}
		printHealthyStatus(cmd)
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("degraded", false, "Show degraded status sample")
	output.AddFlag(statusCmd)
}

func buildStatusSummary(degraded bool) map[string]any {
	services := []map[string]any{
		{"name": "HTTP API", "status": ternaryStatus(!degraded, "online", "online"), "address": "localhost:8080", "uptime": ternary(degraded, "45m", "2h 15m"), "load": ternary(degraded, "8 req/min", "12 req/min")},
		{"name": "TCP Sync", "status": ternaryStatus(!degraded, "online", "error"), "address": "localhost:9090", "uptime": ternary(degraded, "-", "2h 15m"), "load": ternary(degraded, "-", "3 clients")},
		{"name": "UDP Notifications", "status": ternaryStatus(!degraded, "online", "warn"), "address": "localhost:9091", "uptime": ternary(degraded, "45m", "2h 15m"), "load": ternary(degraded, "0 clients", "8 clients")},
		{"name": "gRPC Internal", "status": "online", "address": "localhost:9092", "uptime": ternary(degraded, "45m", "2h 15m"), "load": ternary(degraded, "2 req/min", "5 req/min")},
		{"name": "WebSocket Chat", "status": "online", "address": "localhost:9093", "uptime": ternary(degraded, "45m", "2h 15m"), "load": ternary(degraded, "5 users", "12 users")},
	}

	overall := "healthy"
	issues := []string{}
	if degraded {
		overall = "degraded"
		issues = append(issues, "TCP Sync Server: Port 9090 already in use", "UDP Notifications: No clients registered")
	}

	return map[string]any{
		"overall":   overall,
		"services":  services,
		"database":  map[string]any{"connection": "active", "size": "2.1 MB", "tables": []string{"users", "manga", "user_progress"}, "last_backup": "2024-01-20 12:00:00"},
		"resources": map[string]string{"memory": "45.2 MB / 512 MB (8.8%)", "cpu": "2.3%", "disk": "892 MB / 10 GB available"},
		"issues":    issues,
	}
}

func ternary(condition bool, whenTrue, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func ternaryStatus(ok bool, healthyValue, degradedValue string) string {
	if ok {
		return healthyValue
	}
	return degradedValue
}

func printHealthyStatus(cmd *cobra.Command) {
	cmd.Println("MangaHub Server Status")
	cmd.Println()
	cmd.Println("┌─────────────────────┬──────────┬─────────────────────┬────────────┬──────────────┐")
	cmd.Println("│ Service             │ Status   │ Address             │ Uptime     │ Load         │")
	cmd.Println("├─────────────────────┼──────────┼─────────────────────┼────────────┼──────────────┤")
	cmd.Println("│ HTTP API            │ ✓ Online │ localhost:8080      │ 2h 15m     │ 12 req/min   │")
	cmd.Println("│ TCP Sync            │ ✓ Online │ localhost:9090      │ 2h 15m     │ 3 clients    │")
	cmd.Println("│ UDP Notifications   │ ✓ Online │ localhost:9091      │ 2h 15m     │ 8 clients    │")
	cmd.Println("│ gRPC Internal       │ ✓ Online │ localhost:9092      │ 2h 15m     │ 5 req/min    │")
	cmd.Println("│ WebSocket Chat      │ ✓ Online │ localhost:9093      │ 2h 15m     │ 12 users     │")
	cmd.Println("└─────────────────────┴──────────┴─────────────────────┴────────────┴──────────────┘")
	cmd.Println()
	cmd.Println("Overall System Health: ✓ Healthy")
	cmd.Println()
	cmd.Println("Database:")
	cmd.Println("Connection: ✓ Active")
	cmd.Println("Size: 2.1 MB")
	cmd.Println("Tables: 3 (users, manga, user_progress)")
	cmd.Println("Last backup: 2024-01-20 12:00:00")
	cmd.Println()
	cmd.Println("Memory Usage: 45.2 MB / 512 MB (8.8%)")
	cmd.Println("CPU Usage: 2.3% average")
	cmd.Println("Disk Space: 892 MB / 10 GB available")
}

func printDegradedStatus(cmd *cobra.Command) {
	cmd.Println("MangaHub Server Status")
	cmd.Println()
	cmd.Println("┌─────────────────────┬──────────┬─────────────────────┬────────────┬──────────────┐")
	cmd.Println("│ Service             │ Status   │ Address             │ Uptime     │ Load         │")
	cmd.Println("├─────────────────────┼──────────┼─────────────────────┼────────────┼──────────────┤")
	cmd.Println("│ HTTP API            │ ✓ Online │ localhost:8080      │ 45m        │ 8 req/min    │")
	cmd.Println("│ TCP Sync            │ ✗ Error  │ localhost:9090      │ -          │ -            │")
	cmd.Println("│ UDP Notifications   │ ⚠ Warn   │ localhost:9091      │ 45m        │ 0 clients    │")
	cmd.Println("│ gRPC Internal       │ ✓ Online │ localhost:9092      │ 45m        │ 2 req/min    │")
	cmd.Println("│ WebSocket Chat      │ ✓ Online │ localhost:9093      │ 45m        │ 5 users      │")
	cmd.Println("└─────────────────────┴──────────┴─────────────────────┴────────────┴──────────────┘")
	cmd.Println()
	cmd.Println("Overall System Health: ⚠ Degraded")
	cmd.Println()
	cmd.Println("Issues Detected:")
	cmd.Println("  ✗ TCP Sync Server: Port 9090 already in use")
	cmd.Println("    Solution: Kill process on port 9090 or change port in config")
	cmd.Println()
	cmd.Println("  ⚠ UDP Notifications: No clients registered")
	cmd.Println("    This is normal if no users have subscribed to notifications")
	cmd.Println()
	cmd.Println("Run 'mangahub server health' for detailed diagnostics")
}
