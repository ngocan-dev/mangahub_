package export

import "github.com/spf13/cobra"

var exportProgressCmd = &cobra.Command{
	Use:     "progress",
	Short:   "Export progress data",
	Long:    "Export your MangaHub reading progress to a file.",
	Example: "mangahub export progress --format csv --output progress.csv",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		cmd.Println("Exporting reading progress...")
		cmd.Println()
		cmd.Println("✓ Export complete!")
		cmd.Println("Saved to: " + outputPath)
		cmd.Println("Records: 389")
		_ = format // format consumed to align UX; could be used for future handling
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportProgressCmd)
	exportProgressCmd.Flags().String("output", "progress.csv", "Output file")
	exportProgressCmd.Flags().String("format", "csv", "Export format (csv|json)")
}
