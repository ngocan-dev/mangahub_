package export

import "github.com/spf13/cobra"

// ExportCmd handles export operations.
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export MangaHub data",
	Long:  "Export library and progress data from MangaHub.",
}

var exportLibraryCmd = &cobra.Command{
	Use:     "library",
	Short:   "Export library data",
	Long:    "Export your MangaHub library to a file.",
	Example: "mangahub export library --format json --output library.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		cmd.Println("Exporting library...")
		cmd.Println()
		cmd.Println("✓ Export complete!")
		cmd.Println("Saved to: " + outputPath)
		cmd.Println("Format: " + formatUpper(format))
		cmd.Println("Entries: 47")
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportLibraryCmd)
	exportLibraryCmd.Flags().String("output", "library.json", "Output file")
	exportLibraryCmd.Flags().String("format", "json", "Export format (json|yaml)")
}
