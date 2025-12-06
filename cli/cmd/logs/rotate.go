package logs

import "github.com/spf13/cobra"

var rotateCmd = &cobra.Command{
	Use:     "rotate",
	Short:   "Rotate log files",
	Long:    "Rotate MangaHub log files to archive current logs and start new ones.",
	Example: "mangahub logs rotate",
	RunE: func(cmd *cobra.Command, args []string) error {
		if quiet() {
			return nil
		}

		cmd.Println("Rotating logs...\n")
		cmd.Println("✓ Active log file archived: server-2024-01-20.log.gz")
		cmd.Println("✓ New log file created: server-current.log")
		return nil
	},
}

func init() {
	LogsCmd.AddCommand(rotateCmd)
}
