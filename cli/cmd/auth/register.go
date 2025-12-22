package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"syscall"
	"time"
	"unicode"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AuthCmd is the parent for authentication operations.
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage MangaHub authentication including registration and login.",
}

var authRegisterCmd = NewRegisterCommand("register", "Register a new MangaHub account", "")

// NewRegisterCommand builds a reusable register command.
func NewRegisterCommand(use, short, example string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Example: example,
		RunE:    runAuthRegister,
	}
	cmd.Flags().String("username", "", "Username for the new account")
	cmd.Flags().String("email", "", "Email address for the new account")
	cmd.Flags().String("password", "", "Password for the new account (optional; will prompt if empty)")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("email")
	return cmd
}

func init() {
	AuthCmd.AddCommand(authRegisterCmd)
}

func runAuthRegister(cmd *cobra.Command, args []string) error {
	username, _ := cmd.Flags().GetString("username")
	email, _ := cmd.Flags().GetString("email")

	if err := validateUserInput(username, email); err != nil {
		return handleValidationError(cmd, err)
	}

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		confirm := ""
		var err error
		password, confirm, err = promptPassword(cmd)
		if err != nil {
			return err
		}

		if password != confirm {
			printPasswordMismatch(cmd)
			os.Exit(1)
		}
	}

	if err := validatePassword(password); err != nil {
		printWeakPassword(cmd)
		os.Exit(1)
	}

	cfg := config.ManagerInstance()
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
	resp, err := client.Register(cmd.Context(), username, email, password)
	if err != nil {
		return handleRegisterError(cmd, err, username)
	}

	output.PrintJSON(cmd, resp)
	printRegisterSuccess(cmd, resp)

	return nil
}

func validateUserInput(username, email string) error {
	if username == "" || email == "" {
		return errors.New("both username and email are required")
	}

	emailPattern := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	if !emailPattern.MatchString(email) {
		return errInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errWeakPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errWeakPassword
	}

	return nil
}

func promptPassword(cmd *cobra.Command) (string, string, error) {
	fmt.Fprint(cmd.OutOrStdout(), "Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(cmd.OutOrStdout())
	if err != nil {
		return "", "", err
	}

	fmt.Fprint(cmd.OutOrStdout(), "Confirm password: ")
	confirm, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(cmd.OutOrStdout())
	if err != nil {
		return "", "", err
	}

	return string(password), string(confirm), nil
}

func handleValidationError(cmd *cobra.Command, err error) error {
	switch err {
	case errInvalidEmail:
		printInvalidEmail(cmd)
		os.Exit(1)
	case errWeakPassword:
		printWeakPassword(cmd)
		os.Exit(1)
	}

	return err
}

func handleRegisterError(cmd *cobra.Command, err error, username string) error {
	if apiErr, ok := err.(*api.Error); ok {
		switch apiErr.Code {
		case "username_exists":
			printUsernameExists(cmd, username)
			os.Exit(1)
		case "invalid_email":
			printInvalidEmail(cmd)
			os.Exit(1)
		case "weak_password":
			printWeakPassword(cmd)
			os.Exit(1)
		}
	}

	return err
}

func printUsernameExists(cmd *cobra.Command, username string) {
	cmd.Println("❌ Username already exists")
	cmd.Printf("✗ Registration failed: Username '%s' already exists\n", username)
	cmd.Printf("Try: mangahub auth login --username %s\n", username)
}

func printInvalidEmail(cmd *cobra.Command) {
	cmd.Println("❌ Invalid email format")
	cmd.Println("✗ Registration failed: Invalid email format")
	cmd.Println("Please provide a valid email address")
}

func printWeakPassword(cmd *cobra.Command) {
	cmd.Println("❌ Weak password")
	cmd.Println("✗ Registration failed: Password too weak")
	cmd.Println("Password must be at least 8 characters with mixed case and numbers")
}

func printPasswordMismatch(cmd *cobra.Command) {
	cmd.Println("❌ Passwords do not match")
	cmd.Println("✗ Registration failed: Passwords do not match")
	cmd.Println("Please try again.")
}

func printRegisterSuccess(cmd *cobra.Command, resp *api.RegisterResponse) {
	created := resp.CreatedAt
	if parsed, err := time.Parse(time.RFC3339, resp.CreatedAt); err == nil {
		created = parsed.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	cmd.Println("✓ Account created successfully!")
	cmd.Printf("User ID: %s\n", resp.ID)
	cmd.Printf("Username: %s\n", resp.Username)
	cmd.Printf("Email: %s\n", resp.Email)
	cmd.Printf("Created: %s\n", created)
	cmd.Println()
	cmd.Println("Please login to start using MangaHub:")
	cmd.Printf("mangahub auth login --username %s\n", resp.Username)
}

var (
	errInvalidEmail = errors.New("invalid email")
	errWeakPassword = errors.New("weak password")
)
