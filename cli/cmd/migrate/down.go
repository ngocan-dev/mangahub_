package migrate

import (
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Rollback database migrations",
	Long:    "List migrations in reverse order for rollback planning.",
	Example: "mangahub migrate down",
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

		for i, j := 0, len(migrations)-1; i < j; i, j = i+1, j-1 {
			migrations[i], migrations[j] = migrations[j], migrations[i]
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"migrations": migrations})
			return nil
		}

		if len(migrations) == 0 {
			cmd.Println("No migrations found.")
			return nil
		}

		cmd.Printf("Rollback plan for %d migrations from %s...\n", len(migrations), path)
		for _, migration := range migrations {
			cmd.Printf("- %s\n", filepath.Base(migration))
		}
		cmd.Println("✓ Migration listing complete (rollback using your backend runner).")
		return nil
	},
}

func init() {
	MigrateCmd.AddCommand(downCmd)
	downCmd.Flags().String("path", "backend/db/migrations", "Path to migration files")
	output.AddFlag(downCmd)
}
