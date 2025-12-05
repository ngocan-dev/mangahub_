package progress

import "github.com/spf13/cobra"

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Sync progress",
	Long:    "Synchronize progress with remote MangaHub services.",
	Example: "mangahub progress sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement progress sync
		cmd.Println("Progress sync is not yet implemented.")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("force", false, "Force synchronization")
}
