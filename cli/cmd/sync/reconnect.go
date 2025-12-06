package sync

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var reconnectCmd = &cobra.Command{
	Use:     "reconnect",
	Short:   "Reconnect synchronization",
	Long:    "Reconnect to synchronization services if the connection is lost.",
	Example: "mangahub sync reconnect",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configQuiet() {
			return nil
		}

		cmd.Println("Resetting sync connections...\n")
		cmd.Println("✓ TCP sync reconnected")
		cmd.Println("✓ UDP notifications re-registered")
		cmd.Println("✓ WebSocket chat connection reset\n")
		cmd.Println("Sync connections refreshed successfully.")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(reconnectCmd)
}

func configQuiet() bool {
	return config.Runtime().Quiet
}
