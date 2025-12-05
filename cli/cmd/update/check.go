package update

import "github.com/spf13/cobra"

// UpdateCmd handles CLI update operations.
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Manage MangaHub CLI updates",
	Long:  "Check for and install MangaHub CLI updates.",
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check for updates",
	Long:    "Check if a new version of the MangaHub CLI is available.",
	Example: "mangahub update check",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement update check
		cmd.Println("Update check is not yet implemented.")
		return nil
	},
}

func init() {
	UpdateCmd.AddCommand(checkCmd)
}
