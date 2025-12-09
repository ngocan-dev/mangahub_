package server

import (
	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the server",
	Long:    "Stop the running MangaHub server instance.",
	Example: "mangahub server stop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Stopping MangaHub Servers...")
		cmd.Println()

		for _, component := range serverstate.Components() {
			cmd.Println("✓ " + component.StopLabel)
		}

		cmd.Println()
		cmd.Println("All services stopped.")
		serverstate.MarkStopped()
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(stopCmd)
}
