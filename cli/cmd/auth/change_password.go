package auth

import "github.com/spf13/cobra"

var changePasswordCmd = &cobra.Command{
	Use:     "change-password",
	Short:   "Change the account password",
	Long:    "Update the password for the currently authenticated MangaHub account.",
	Example: "mangahub auth change-password --current old --new newpass",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement password change logic
		cmd.Println("Password change is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(changePasswordCmd)
	changePasswordCmd.Flags().String("current", "", "Current password")
	changePasswordCmd.Flags().String("new", "", "New password")
}
