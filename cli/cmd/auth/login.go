package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = NewLoginCommand("login", "Login to your MangaHub account", "")

// NewLoginCommand builds a reusable login command.
func NewLoginCommand(use, short, example string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Example: example,
		RunE:    runAuthLogin,
	}
	cmd.Flags().String("username", "", "Username for login")
	cmd.Flags().String("email", "", "Email for login")
	cmd.Flags().String("password", "", "Password for login (optional; will prompt if empty)")
	return cmd
}

func init() {
	AuthCmd.AddCommand(loginCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	username, _ := cmd.Flags().GetString("username")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	if username == "" && email == "" {
		return errors.New("either --username or --email is required")
	}

	if username != "" && email != "" {
		return errors.New("Please use either --username OR --email")
	}

	if password == "" {
		fmt.Fprint(cmd.OutOrStdout(), "Password: ")
		rawPassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(cmd.OutOrStdout())
		if err != nil {
			return err
		}
		password = string(rawPassword)
	}

	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
	resp, err := client.Login(cmd.Context(), username, email, password)
	if err != nil {
		handleLoginError(cmd, err, username, email)
		return err
	}

	if err := cfg.UpdateSession(resp.Token, resp.ExpiresAt, resp.User.Username, resp.User.Permissions, resp.User.Settings.Autosync, resp.User.Settings.Notifications); err != nil {
		return err
	}

	if config.Runtime().Verbose {
		output.PrintJSON(cmd, resp)
		cmd.Printf("Config saved to %s\n", cfg.Path)
	}

	printLoginSuccess(cmd, resp, config.Runtime().Quiet)
	return nil
}

func handleLoginError(cmd *cobra.Command, err error, username, email string) {
	if apiErr, ok := err.(*api.Error); ok {
		switch apiErr.Code {
		case "invalid_credentials":
			cmd.Println("❌ 1. Invalid credentials")
			cmd.Println("✗ Login failed: Invalid credentials")
			cmd.Println("Check your username and password")
			os.Exit(1)
		case "account_not_found":
			cmd.Println("❌ 2. Account not found")
			cmd.Println("✗ Login failed: Account not found")
			if username != "" {
				cmd.Printf("Try: mangahub auth register --username %s --email john@example.com\n", username)
			} else {
				cmd.Printf("Try: mangahub auth register --username johndoe --email %s\n", email)
			}
			os.Exit(1)
		}
	}

	cmd.Println("❌ 3. Server connection error")
	cmd.Println("✗ Login failed: Server connection error")
	cmd.Println("Check server status: mangahub server status")
	os.Exit(1)
}

func printLoginSuccess(cmd *cobra.Command, resp *api.LoginResponse, quiet bool) {
	if quiet {
		cmd.Println("✓ Login successful!")
		return
	}

	expTime := resp.ExpiresAt
	if parsed, err := time.Parse(time.RFC3339, resp.ExpiresAt); err == nil {
		expTime = parsed.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	permissions := strings.Join(resp.User.Permissions, ", ")
	autosync := "disabled"
	if resp.User.Settings.Autosync {
		autosync = "enabled"
	}
	notifications := "disabled"
	if resp.User.Settings.Notifications {
		notifications = "enabled"
	}

	cmd.Println("✓ Login successful!")
	cmd.Printf("Welcome back, %s!\n\n", resp.User.Username)
	cmd.Println("Session Details:")
	cmd.Printf("Token expires: %s (24 hours)\n", expTime)
	cmd.Printf("Permissions: %s\n", permissions)
	cmd.Printf("Auto-sync: %s\n", autosync)
	cmd.Printf("Notifications: %s\n\n", notifications)
	cmd.Println("Ready to use MangaHub! Try:")
	cmd.Println("mangahub manga search \"your favorite manga\"")
}
