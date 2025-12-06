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
	Example: "mangahub export library --output library.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement library export
		cmd.Println("Library export is not yet implemented.")
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportLibraryCmd)
	exportLibraryCmd.Flags().String("output", "library.json", "Output file")
}
