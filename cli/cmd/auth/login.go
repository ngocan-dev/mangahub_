package auth

import "github.com/spf13/cobra"

var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to MangaHub",
	Long:    "Authenticate with MangaHub using username/email and password.",
	Example: "mangahub auth login --username alice --password secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement login logic
		cmd.Println("Login is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("username", "", "Username for login")
	loginCmd.Flags().String("password", "", "Password for login")
}
