package library

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List library entries",
	Long:    "List all manga saved in your MangaHub library.",
	Example: "mangahub library list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement library listing
		cmd.Println("Library listing is not yet implemented.")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(listCmd)
}
