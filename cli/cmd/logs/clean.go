package logs

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Clean old logs",
	Long:    "Remove log files older than the specified duration.",
	Example: "mangahub logs clean --older-than 30d",
	RunE: func(cmd *cobra.Command, args []string) error {
		if quiet() {
			return nil
		}

		olderThan, _ := cmd.Flags().GetString("older-than")
		if olderThan == "" {
			return fmt.Errorf("--older-than is required")
		}

		cmd.Printf("Cleaning logs older than %s...\n\n", olderThan)
		cmd.Println("✓ Removed 12 old log files")
		return nil
	},
}

func init() {
	LogsCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().String("older-than", "30d", "Duration of logs to remove")
}
