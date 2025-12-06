package export

import "github.com/spf13/cobra"

var exportProgressCmd = &cobra.Command{
	Use:     "progress",
	Short:   "Export progress data",
	Long:    "Export your MangaHub reading progress to a file.",
	Example: "mangahub export progress --output progress.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement progress export
		cmd.Println("Progress export is not yet implemented.")
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportProgressCmd)
	exportProgressCmd.Flags().String("output", "progress.json", "Output file")
}
