package sync

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show sync status",
	Long:    "Display the status of synchronization connections.",
	Example: "mangahub sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement sync status
		cmd.Println("Sync status is not yet implemented.")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(statusCmd)
}
