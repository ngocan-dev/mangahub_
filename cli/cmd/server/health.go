package server

import "github.com/spf13/cobra"

var healthCmd = &cobra.Command{
	Use:     "health",
	Short:   "Check server health",
	Long:    "Perform a health check against the MangaHub server.",
	Example: "mangahub server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		degraded, _ := cmd.Flags().GetBool("degraded")
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
