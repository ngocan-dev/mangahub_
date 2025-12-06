package auth

import (
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Clear stored authentication data",
	Long:    "Remove cached authentication credentials from the local environment.",
	Example: "mangahub auth clear",
	RunE:    runAuthClear,
}

func init() {
	AuthCmd.AddCommand(clearCmd)
}

func runAuthClear(cmd *cobra.Command, args []string) error {
	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	cmd.Println("Clearing authentication data...")
	cmd.Println()

	// Clear token and session
	if err := cfg.ClearSession(); err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	cmd.Println("✓ Authentication token removed")
	cmd.Println("✓ User session cleared")
	cmd.Println("✓ Sync connections terminated")
	cmd.Println("✓ Cache cleared")

	cmd.Println()
	cmd.Println("You are now logged out. To continue using MangaHub:")
	cmd.Println("  mangahub auth login --username <your-username>")
	cmd.Println()
	cmd.Println("Or register a new account:")
	cmd.Println("  mangahub auth register --username <username> --email <email>")

	return nil
}
