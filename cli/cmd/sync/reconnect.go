package sync

import "github.com/spf13/cobra"

var reconnectCmd = &cobra.Command{
	Use:     "reconnect",
	Short:   "Reconnect synchronization",
	Long:    "Reconnect to synchronization services if the connection is lost.",
	Example: "mangahub sync reconnect",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement sync reconnect
		cmd.Println("Sync reconnect is not yet implemented.")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(reconnectCmd)
}
