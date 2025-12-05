package manga

import "github.com/spf13/cobra"

// MangaCmd groups manga-related commands.
var MangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Interact with manga metadata",
	Long:  "Search and retrieve manga information from MangaHub services.",
}

var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search for manga titles",
	Long:    "Search MangaHub for manga titles using keywords and optional filters.",
	Example: "mangahub manga search --query 'One Piece'",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement manga search
		cmd.Println("Manga search is not yet implemented.")
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(searchCmd)
	searchCmd.Flags().String("query", "", "Query string to search manga")
	searchCmd.Flags().String("genre", "", "Filter by genre")
}
