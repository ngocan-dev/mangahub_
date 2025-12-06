package auth

import (
	"errors"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// loginCmd authenticates a user and stores the token in the config file.
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to MangaHub",
	Long:    "Authenticate with MangaHub using a username and save the access token for future requests.",
	Example: "mangahub auth login --username alice",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		if username == "" {
			return errors.New("--username is required")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		token, payload, err := client.Login(cmd.Context(), username)
		if err != nil {
			return err
		}

		if err := cfg.UpdateToken(token); err != nil {
			return err
		}

		output.PrintJSON(cmd, payload)
		output.PrintSuccess(cmd, "Login successful")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("username", "", "Username for login")
	loginCmd.MarkFlagRequired("username")
}
