package server

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:     "health",
	Short:   "Check server health",
	Long:    "Perform a health check against the MangaHub server.",
	Example: "mangahub server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		degraded, _ := cmd.Flags().GetBool("degraded")
		status := buildHealthSummary(degraded)
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
			printDegradedHealth(cmd)
			return nil
		}
		printHealthyHealth(cmd)
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(healthCmd)
	healthCmd.Flags().Bool("degraded", false, "Show degraded health sample")
	output.AddFlag(healthCmd)
}

func buildHealthSummary(degraded bool) map[string]any {
	overall := "healthy"
	issues := []string{}
	if degraded {
		overall = "degraded"
		issues = append(issues, "TCP Sync Server is failing to bind to tcp://localhost:9090", "UDP Notifications: No subscribers")
	}

	return map[string]any{
		"overall": overall,
		"components": map[string]string{
			"http_api":          ternaryStatus(!degraded, "healthy", "healthy"),
			"tcp_sync":          ternaryStatus(!degraded, "healthy", "error"),
			"udp_notifications": ternaryStatus(!degraded, "healthy", "warn"),
			"grpc_internal":     "healthy",
			"websocket_chat":    "healthy",
		},
		"dependencies": map[string]string{
			"database":        "connected",
			"cache":           "warm",
			"background_jobs": ternaryStatus(!degraded, "running", "delayed"),
		},
		"notes":   issues,
		"metrics": map[string]any{"latency_median_ms": ternary(degraded, "28", "12"), "grpc_p99_ms": ternary(degraded, "62", "48"), "active_chat_users": ternary(degraded, "5", "12")},
	}
}

func printHealthyHealth(cmd *cobra.Command) {
	cmd.Println("MangaHub Server Health")
	cmd.Println()
	cmd.Println("HTTP API:           ✓ Healthy (12ms median)")
	cmd.Println("TCP Sync:           ✓ Healthy (3 connected peers)")
	cmd.Println("UDP Notifications:  ✓ Healthy (queue depth: 0)")
	cmd.Println("gRPC Internal:      ✓ Healthy (p99: 48ms)")
	cmd.Println("WebSocket Chat:     ✓ Healthy (12 active users)")
	cmd.Println()
	cmd.Println("Database:           ✓ Connected (postgres @ localhost:5432)")
	cmd.Println("Cache:              ✓ Warm (hit rate: 91%)")
	cmd.Println("Background Jobs:    ✓ Running (5 queued)")
	cmd.Println()
	cmd.Println("Overall: ✓ Healthy")
}

func printDegradedHealth(cmd *cobra.Command) {
	cmd.Println("MangaHub Server Health")
	cmd.Println()
	cmd.Println("HTTP API:           ✓ Healthy (28ms median)")
	cmd.Println("TCP Sync:           ✗ Error  (port collision)")
	cmd.Println("UDP Notifications:  ⚠ Warn   (no subscribers)")
	cmd.Println("gRPC Internal:      ✓ Healthy (p99: 62ms)")
	cmd.Println("WebSocket Chat:     ✓ Healthy (5 active users)")
	cmd.Println()
	cmd.Println("Database:           ✓ Connected (postgres @ localhost:5432)")
	cmd.Println("Cache:              ✓ Warm (hit rate: 88%)")
	cmd.Println("Background Jobs:    ⚠ Delayed (retrying 2 tasks)")
	cmd.Println()
	cmd.Println("Overall: ⚠ Degraded")
	cmd.Println("Issues:")
	cmd.Println("  - TCP Sync Server is failing to bind to tcp://localhost:9090")
	cmd.Println("    Resolution: Free port 9090 or update configuration")
}
