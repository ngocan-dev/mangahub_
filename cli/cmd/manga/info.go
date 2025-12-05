package manga

import "github.com/spf13/cobra"

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show manga details",
	Long:    "Retrieve detailed information about a specific manga by ID or slug.",
	Example: "mangahub manga info --id 123",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement manga info retrieval
		cmd.Println("Manga info is not yet implemented.")
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(infoCmd)
	infoCmd.Flags().String("id", "", "Manga identifier")
	infoCmd.Flags().String("slug", "", "Manga slug")
}
