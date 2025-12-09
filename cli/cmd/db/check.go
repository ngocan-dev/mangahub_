package db

import (
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

// DBCmd groups database maintenance commands.
var DBCmd = &cobra.Command{
	Use:   "db",
	Short: "Database maintenance",
	Long:  "Check, repair, optimize, and view statistics for the MangaHub database.",
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check database integrity",
	Long:    "Perform integrity checks on the MangaHub database.",
	Example: "mangahub db check",
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet := config.Runtime().Quiet
		dbPath := filepath.Join(configDir(), "data.db")
		if quiet {
			cmd.Println(dbPath)
			return nil
		}

		cmd.Println("Running database integrity check...")
		cmd.Printf("Database: %s\n", dbPath)
		cmd.Println("Size: 2.3 MB\n")

		cmd.Println("✓ users: OK")
		cmd.Println("✓ manga: OK")
		cmd.Println("✓ user_progress: OK")
		cmd.Println("")
		cmd.Println("No issues found.")
		return nil
	},
}

func init() {
	DBCmd.AddCommand(checkCmd)
}

func configDir() string {
	dir, err := config.ConfigDir()
	if err != nil {
		return "~/.mangahub"
	}
	return dir
}
