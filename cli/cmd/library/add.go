package library

import "github.com/spf13/cobra"

// LibraryCmd manages library entries.
var LibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage your MangaHub library",
	Long:  "Add, remove, and update manga entries in your personal library.",
}

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a manga to the library",
	Long:    "Add a manga to your library by ID or slug with optional status.",
	Example: "mangahub library add --id 123 --status reading",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement library add
		cmd.Println("Library add is not yet implemented.")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(addCmd)
	addCmd.Flags().String("id", "", "Manga identifier")
	addCmd.Flags().String("status", "", "Reading status")
}
