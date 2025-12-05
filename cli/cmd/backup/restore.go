package backup

import "github.com/spf13/cobra"

var restoreCmd = &cobra.Command{
	Use:     "restore",
	Short:   "Restore a backup",
	Long:    "Restore MangaHub data from a backup archive.",
	Example: "mangahub backup restore --input backup.zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement backup restore
		cmd.Println("Backup restore is not yet implemented.")
		return nil
	},
}

func init() {
	BackupCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().String("input", "backup.zip", "Backup file to restore")
}
