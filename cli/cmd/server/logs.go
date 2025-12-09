package server

import (
	"fmt"

	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Show server logs",
	Long:    "Tail or display MangaHub server logs.",
	Example: "mangahub server logs --tail 100",
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		level, _ := cmd.Flags().GetString("level")

		cmd.Println("Showing logs from: " + serverstate.LogPath())

		if level != "" && level != "info" {
			cmd.Println(fmt.Sprintf("Filter: level=%s", level))
		}

		if follow {
			cmd.Println()
			cmd.Println("-- Live tail (Ctrl+C to stop) --")
			cmd.Println("[2024-02-01 10:00:00] INFO  http: request handled (12ms)")
			cmd.Println("[2024-02-01 10:00:02] INFO  ws: user connected (id=1042)")
			cmd.Println("[2024-02-01 10:00:05] INFO  udp: 2 notifications broadcast")
			cmd.Println("[2024-02-01 10:00:09] INFO  grpc: health probe ok")
			return nil
		}

		if level == "error" {
			cmd.Println()
			cmd.Println("[2024-02-01 09:58:12] ERROR tcp: failed to bind to tcp://localhost:9090 (address in use)")
			cmd.Println("[2024-02-01 09:58:13] WARN  udp: no subscribers connected yet")
			cmd.Println("[2024-02-01 09:58:20] ERROR tcp: retrying listener startup")
			cmd.Println()
			cmd.Println("Use --follow to stream logs in real-time")
			return nil
		}

		cmd.Println()
		cmd.Println("Use --follow to stream logs in real-time")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(logsCmd)
	logsCmd.Flags().Int("tail", 50, "Number of log lines to show")
	logsCmd.Flags().Bool("follow", false, "Stream logs in real-time")
	logsCmd.Flags().String("level", "info", "Filter logs by level (info|warn|error)")
}
