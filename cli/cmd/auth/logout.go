package auth

import (
	"errors"
	"os"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from MangaHub",
	Long:    "Clear local authentication tokens and logout from MangaHub.",
	Example: "mangahub auth logout",
	RunE:    runAuthLogout,
}

// NewLogoutCommand builds a reusable logout command.
func NewLogoutCommand(use, short, long, example string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: example,
		RunE:    runAuthLogout,
	}
}

func init() {
	AuthCmd.AddCommand(logoutCmd)
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	if cfg.Data.Token == "" && cfg.Data.Auth.Username == "" {
		cmd.Println("✗ You are not logged in.")
		cmd.Println("Nothing to do.")
		os.Exit(1)
	}

	if err := cfg.ClearSession(); err != nil {
		return err
	}

	if config.Runtime().Verbose {
		cmd.Printf("Config cleared at %s\n", cfg.Path)
	}

	if config.Runtime().Quiet {
		cmd.Println("✓ Logged out successfully.")
		return nil
	}

	cmd.Println("✓ Logged out successfully.")
	cmd.Println("Your authentication token has been removed.")
	return nil
}
