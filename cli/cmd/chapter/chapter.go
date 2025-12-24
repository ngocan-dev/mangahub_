package chapter

import "github.com/spf13/cobra"

// ChapterCmd groups chapter-related commands.
var ChapterCmd = &cobra.Command{
	Use:   "chapter",
	Short: "Browse and read chapters",
	Long:  "List and read chapter content from MangaHub.",
}
