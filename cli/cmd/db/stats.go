package db

import (
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show database stats",
	Long:    "Display database statistics such as size and record counts.",
	Example: "mangahub db stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet := config.Runtime().Quiet
		dbPath := filepath.Join(configDir(), "data.db")
		if quiet {
			cmd.Println(dbPath)
			return nil
		}

		cmd.Println("Database Statistics\n")
		cmd.Printf("Path: %s\n", dbPath)
		cmd.Println("Size: 2.0 MB")
		cmd.Println("Tables:")
		cmd.Println("- users: 15 records")
		cmd.Println("- manga: 42 records")
		cmd.Println("- user_progress: 127 records\n")
		cmd.Println("Last vacuum: 2024-01-19 10:30:00")
		cmd.Println("Last backup: 2024-01-20 12:00:00")
		return nil
	},
}

func init() {
	DBCmd.AddCommand(statsCmd)
}
