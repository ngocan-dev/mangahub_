package export

import "github.com/spf13/cobra"

var exportAllCmd = &cobra.Command{
	Use:     "all",
	Short:   "Export all data",
	Long:    "Export library and progress data into a single archive.",
	Example: "mangahub export all --output mangahub_export.zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement full export
		cmd.Println("Full export is not yet implemented.")
		return nil
	},
}

func init() {
	ExportCmd.AddCommand(exportAllCmd)
	exportAllCmd.Flags().String("output", "mangahub_export.zip", "Output archive")
}
