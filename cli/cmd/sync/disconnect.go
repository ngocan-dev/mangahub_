package sync

import "github.com/spf13/cobra"

var disconnectCmd = &cobra.Command{
	Use:     "disconnect",
	Short:   "Disconnect synchronization",
	Long:    "Disconnect synchronization services and cleanup resources.",
	Example: "mangahub sync disconnect",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement sync disconnect
		cmd.Println("Sync disconnect is not yet implemented.")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(disconnectCmd)
}
