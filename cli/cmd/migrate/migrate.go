package migrate

import "github.com/spf13/cobra"

// MigrateCmd manages database migrations.
var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Apply or rollback database migrations for the MangaHub backend.",
}
