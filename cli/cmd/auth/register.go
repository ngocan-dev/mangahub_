package auth

import (
	"errors"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// AuthCmd is the parent for authentication operations.
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage MangaHub authentication including registration and login.",
}

// registerCmd handles account registration.
var registerCmd = &cobra.Command{
	Use:     "register",
	Short:   "Register a new MangaHub account",
	Long:    "Create a new MangaHub account using a username and email address.",
	Example: "mangahub auth register --username alice --email alice@example.com",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")

		if username == "" || email == "" {
			return errors.New("both --username and --email are required")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		payload, err := client.Register(cmd.Context(), username, email)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, payload)
		output.PrintSuccess(cmd, "Registration successful")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("username", "", "Username for the new account")
	registerCmd.Flags().String("email", "", "Email address for the new account")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")
}
