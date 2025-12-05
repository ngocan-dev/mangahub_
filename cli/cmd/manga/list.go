package manga

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List manga",
	Long:    "List manga available in MangaHub with optional pagination.",
	Example: "mangahub manga list --page 1 --page-size 20",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement manga listing
		cmd.Println("Manga listing is not yet implemented.")
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(listCmd)
	listCmd.Flags().Int("page", 1, "Page number")
	listCmd.Flags().Int("page-size", 20, "Page size")
}
