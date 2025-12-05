package library

import "github.com/spf13/cobra"

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a library entry",
	Long:    "Update progress or metadata for a manga in your library.",
	Example: "mangahub library update --id 123 --status completed",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement library update
		cmd.Println("Library update is not yet implemented.")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("id", "", "Manga identifier")
	updateCmd.Flags().String("status", "", "Updated status")
	updateCmd.Flags().Int("chapter", 0, "Current chapter")
}
