package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show authentication status",
	Long:    "Display the current authentication status and active profile.",
	Example: "mangahub auth status",
	RunE:    runAuthStatus,
}

func init() {
	AuthCmd.AddCommand(statusCmd)
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	runtime := config.Runtime()
	if runtime.Quiet {
		if !isSessionValid(cfg.Data) {
			cmd.Println("not-logged-in")
			os.Exit(1)
		}
		cmd.Println("logged-in")
		return nil
	}

	if runtime.Verbose {
		output.PrintJSON(cmd, cfg.Data)
		cmd.Printf("Config loaded from %s\n", cfg.Path)
	}

	if cfg.Data.Token == "" {
		printNotLoggedIn(cmd)
		os.Exit(1)
	}

	if isExpired(cfg.Data.ExpiresAt) {
		printSessionExpired(cmd, cfg.Data.Username)
		os.Exit(1)
	}

	printLoggedInStatus(cmd, cfg.Data)
	return nil
}

func isSessionValid(cfg config.Config) bool {
	if cfg.Token == "" {
		return false
	}
	return !isExpired(cfg.ExpiresAt)
}

func isExpired(exp string) bool {
	if exp == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, exp)
	if err != nil {
		return true
	}
	return time.Now().After(t)
}

func printNotLoggedIn(cmd *cobra.Command) {
	cmd.Println("✗ Not logged in.")
	cmd.Println("Please login using:")
	cmd.Println("mangahub auth login --username <username>")
}

func printSessionExpired(cmd *cobra.Command, username string) {
	if username == "" {
		username = "johndoe"
	}
	cmd.Println("✗ Session expired.")
	cmd.Println("Please login again:")
	cmd.Printf("mangahub auth login --username %s\n", username)
}

func printLoggedInStatus(cmd *cobra.Command, cfg config.Config) {
	expires := cfg.ExpiresAt
	if parsed, err := time.Parse(time.RFC3339, cfg.ExpiresAt); err == nil {
		expires = parsed.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	permissions := strings.Join(cfg.Permissions, ", ")

	autosync := "disabled"
	if cfg.Settings.Autosync {
		autosync = "enabled"
	}
	notifications := "disabled"
	if cfg.Settings.Notifications {
		notifications = "enabled"
	}

	cmd.Printf("✓ You are logged in as %s\n\n", cfg.Username)
	cmd.Println("Session Information:")
	cmd.Printf("Token expires: %s\n", expires)
	cmd.Printf("Permissions: %s\n", permissions)
	cmd.Printf("Auto-sync: %s\n", autosync)
	cmd.Printf("Notifications: %s\n", notifications)
}
