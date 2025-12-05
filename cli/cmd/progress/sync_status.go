package progress

import "github.com/spf13/cobra"

var syncStatusCmd = &cobra.Command{
	Use:     "sync-status",
	Short:   "Check sync status",
	Long:    "View the status of the last or current progress synchronization.",
	Example: "mangahub progress sync-status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement sync status
		cmd.Println("Progress sync status is not yet implemented.")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(syncStatusCmd)
}
