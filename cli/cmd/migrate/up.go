package migrate

import (
	"fmt"
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Apply database migrations",
	Long:    "Apply all pending database migrations from the migration directory.",
	Example: "mangahub migrate up",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		path, _ := cmd.Flags().GetString("path")
		migrations, err := loadMigrations(path)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"migrations": migrations})
			return nil
		}

		if len(migrations) == 0 {
			cmd.Println("No migrations found.")
			return nil
		}

		cmd.Printf("Applying %d migrations from %s...\n", len(migrations), path)
		for _, migration := range migrations {
			cmd.Printf("- %s\n", filepath.Base(migration))
		}
		cmd.Println("✓ Migration listing complete (apply migrations using your backend runner).")
		return nil
	},
}

func init() {
	MigrateCmd.AddCommand(upCmd)
	upCmd.Flags().String("path", "backend/db/migrations", "Path to migration files")
	output.AddFlag(upCmd)
}
