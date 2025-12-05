package auth

import "github.com/spf13/cobra"

// AuthCmd is the parent for authentication operations.
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage MangaHub authentication including registration and login.",
}

var registerCmd = &cobra.Command{
	Use:     "register",
	Short:   "Register a new MangaHub account",
	Long:    "Create a new MangaHub account using a username, email, and password.",
	Example: "mangahub auth register --username alice --email alice@example.com --password secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement registration logic
		cmd.Println("User registration is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("username", "", "Username for the new account")
	registerCmd.Flags().String("email", "", "Email address for the new account")
	registerCmd.Flags().String("password", "", "Password for the new account")
}
