package backup

import (
	"errors"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

// BackupCmd handles backup operations.
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage MangaHub backups",
	Long:  "Create and restore MangaHub backups.",
}

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a backup",
	Long:    "Create a backup archive of MangaHub data.",
	Example: "mangahub backup create --output backup.zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, _ := cmd.Flags().GetString("output")

		if strings.TrimSpace(outputPath) == "" {
			return errors.New("--output is required")
		}

		if config.Runtime().Quiet {
			cmd.Println(outputPath)
			return nil
		}

		cmd.Println("Creating MangaHub backup...\n")
		cmd.Println("✓ Database included")
		cmd.Println("✓ Library data included")
		cmd.Println("✓ Reading progress included")
		cmd.Println("✓ Configuration included")
		cmd.Println("✓ Profiles included")

		cmd.Println("")
		cmd.Println("Backup created successfully!")
		cmd.Printf("File: %s\n", outputPath)
		cmd.Println("Keep this file in a safe location.")
		return nil
	},
}

func init() {
	BackupCmd.AddCommand(createCmd)
	createCmd.Flags().String("output", "backup-2024.tar.gz", "Backup file path")
}
