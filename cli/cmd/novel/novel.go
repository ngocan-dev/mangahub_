package novel

import "github.com/spf13/cobra"

// NovelCmd groups novel-related commands.
var NovelCmd = &cobra.Command{
	Use:   "novel",
	Short: "Interact with novel metadata",
	Long:  "Search and retrieve novel information from MangaHub services.",
}
