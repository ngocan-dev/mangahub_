package auth

import "github.com/spf13/cobra"

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from MangaHub",
	Long:    "Clear local authentication tokens and logout from MangaHub.",
	Example: "mangahub auth logout",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement logout logic
		cmd.Println("Logout is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(logoutCmd)
}
