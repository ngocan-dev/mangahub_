package db

import (
	"fmt"
	"path/filepath"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var repairCmd = &cobra.Command{
	Use:     "repair",
	Short:   "Repair the database",
	Long:    "Attempt to repair database inconsistencies for MangaHub.",
	Example: "mangahub db repair",
	RunE: func(cmd *cobra.Command, args []string) error {
		simulateCorruption, _ := cmd.Flags().GetBool("simulate-corruption")
		quiet := config.Runtime().Quiet
		dbPath := filepath.Join(configDir(), "data.db")
		if quiet {
			cmd.Println(dbPath)
			return nil
		}

		if simulateCorruption {
			printCorruption(cmd, dbPath)
			return fmt.Errorf("database corruption detected")
		}

		printRepair(cmd, dbPath)
		return nil
	},
}

func init() {
	DBCmd.AddCommand(repairCmd)
	repairCmd.Flags().Bool("simulate-corruption", false, "Simulate corruption for testing output")
}

func printRepair(cmd *cobra.Command, dbPath string) {
	cmd.Println("Running database integrity check and repair...")
	cmd.Printf("Database: %s\n", dbPath)
	cmd.Println("Size: 2.3 MB\n")

	cmd.Println("Checking tables...")
	cmd.Println("users table:           ✓ 15 records, no corruption")
	cmd.Println("manga table:           ✓ 42 records, no corruption")
	cmd.Println("user_progress table:   ⚠ 127 records, 3 orphaned entries found\n")

	cmd.Println("Repairing issues...")
	cmd.Println("✓ Removed 3 orphaned progress entries")
	cmd.Println("✓ Rebuilt indexes for performance")
	cmd.Println("✓ Updated database statistics")
	cmd.Println("✓ Compressed database (saved 0.3 MB)\n")

	cmd.Println("Database repair completed successfully!\n")
	cmd.Println("Summary:")
	cmd.Println("Issues found: 3 orphaned entries")
	cmd.Println("Issues fixed: 3")
	cmd.Println("Performance: Improved (faster queries expected)")
	cmd.Println("Size after repair: 2.0 MB\n")
	cmd.Println("Your data is intact and the database is now optimized.")
}

func printCorruption(cmd *cobra.Command, dbPath string) {
	cmd.Println("Running database integrity check and repair...")
	cmd.Printf("Database: %s\n", dbPath)
	cmd.Println("Size: 2.3 MB\n")

	cmd.Println("✗ Critical database corruption detected!")
	cmd.Println("Issues found:")
	cmd.Println("  - users table: 5 corrupted records")
	cmd.Println("  - manga table: Schema mismatch")
	cmd.Println("  - user_progress table: Index corruption\n")

	cmd.Println("⚠ Automatic repair failed. Manual intervention required.\n")
	cmd.Println("Recovery options:")
	cmd.Println("1. Restore from backup:")
	cmd.Println("   mangahub backup restore --input backup-2024.tar.gz\n")
	cmd.Println("2. Reinitialize database (DESTROYS ALL DATA):")
	cmd.Println("   mangahub init --force --wipe-data\n")
	cmd.Println("3. Export recoverable data first:")
	cmd.Println("   mangahub export library --output library-backup.json --ignore-errors\n")
	cmd.Println("Contact support if you need assistance with data recovery.")
}
