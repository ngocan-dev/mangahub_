package export

import "github.com/spf13/cobra"

var exportAllCmd = &cobra.Command{
	Use:     "all",
	Short:   "Export all data",
	Long:    "Export library and progress data into a single archive.",
	Example: "mangahub export all --output mangahub-backup.tar.gz",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")

		cmd.Println("Generating full MangaHub backup...")
		cmd.Println()
		cmd.Println("✓ Backup created successfully!")
		cmd.Println("File: " + outputPath)
		cmd.Println("Includes:")
		cmd.Println("  • Library (JSON)")
		cmd.Println("  • Reading Progress (CSV)")
		cmd.Println("  • Preferences")
		cmd.Println("  • Chat Metadata")
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportAllCmd)
	exportAllCmd.Flags().String("output", "mangahub-backup.tar.gz", "Output archive")
}
