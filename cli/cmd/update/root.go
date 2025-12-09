package update

import "github.com/spf13/cobra"

// UpdateCmd groups CLI update commands.
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install MangaHub CLI updates",
	Long:  "Manage MangaHub CLI updates, including checking for new releases and installing them.",
}
