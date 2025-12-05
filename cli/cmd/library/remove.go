package library

import "github.com/spf13/cobra"

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a manga from the library",
	Long:    "Delete a manga from your library using its identifier.",
	Example: "mangahub library remove --id 123",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement library removal
		cmd.Println("Library removal is not yet implemented.")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(removeCmd)
	removeCmd.Flags().String("id", "", "Manga identifier")
}
