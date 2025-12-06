package auth

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var changePasswordCmd = &cobra.Command{
	Use:     "change-password",
	Short:   "Change the account password",
	Long:    "Update the password for the currently authenticated MangaHub account.",
	Example: "mangahub auth change-password",
	RunE:    runAuthChangePassword,
}

func init() {
	AuthCmd.AddCommand(changePasswordCmd)
}

func runAuthChangePassword(cmd *cobra.Command, args []string) error {
	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	if cfg.Data.Token == "" || isExpired(cfg.Data.ExpiresAt) {
		cmd.Println("✗ You must be logged in to change your password.")
		cmd.Println("Please login first.")
		os.Exit(1)
	}

	if config.Runtime().Verbose {
		cmd.Printf("Using config at %s\n", cfg.Path)
	}

	currentPassword, err := promptHidden(cmd, "Current password: ")
	if err != nil {
		return err
	}

	newPassword, err := promptHidden(cmd, "New password: ")
	if err != nil {
		return err
	}

	confirmPassword, err := promptHidden(cmd, "Confirm new password: ")
	if err != nil {
		return err
	}

	if newPassword != confirmPassword {
		cmd.Println("✗ Password change failed: New passwords do not match")
		cmd.Println("Please try again.")
		os.Exit(1)
	}

	client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
	resp, err := client.ChangePassword(cmd.Context(), currentPassword, newPassword)
	if err != nil {
		handleChangePasswordError(cmd, err)
		return err
	}

	output.PrintJSON(cmd, resp)
	cmd.Println("✓ Password changed successfully!")
	cmd.Println("Your new password is now active.")
	return nil
}

func promptHidden(cmd *cobra.Command, label string) (string, error) {
	fmt.Fprint(cmd.OutOrStdout(), label)
	value, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(cmd.OutOrStdout())
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func handleChangePasswordError(cmd *cobra.Command, err error) {
	if apiErr, ok := err.(*api.Error); ok {
		switch apiErr.Code {
		case "incorrect_current_password":
			cmd.Println("✗ Password change failed: Incorrect current password")
			cmd.Println("Please try again.")
			os.Exit(1)
		case "password_mismatch":
			cmd.Println("✗ Password change failed: New passwords do not match")
			cmd.Println("Please try again.")
			os.Exit(1)
		case "password_too_weak", "weak_password":
			cmd.Println("✗ Password change failed: Password too weak")
			cmd.Println("Password must be at least 8 characters with mixed case and numbers")
			os.Exit(1)
		}
	}

	cmd.Println("✗ Password change failed: Server connection error")
	cmd.Println("Check server status: mangahub server status")
	os.Exit(1)
}
