package user

import (
	"github.com/ngocan-dev/mangahub_/cli/cmd/auth"
	"github.com/spf13/cobra"
)

// UserCmd groups user-related commands.
var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage MangaHub user accounts",
	Long:  "Register and inspect MangaHub user accounts.",
}

func init() {
	registerCmd := auth.NewRegisterCommand("register", "Register a new MangaHub account", "mangahub user register --username <name> --email <email> --password <password>")
	infoCmd := auth.NewStatusCommand("info", "Show user account info", "Show the current account profile and session details.", "mangahub user info")

	UserCmd.AddCommand(registerCmd)
	UserCmd.AddCommand(infoCmd)
}
