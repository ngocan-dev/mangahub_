package progress

import (
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var syncStatusCmd = &cobra.Command{
	Use:     "sync-status",
	Short:   "Check sync status",
	Long:    "View the status of the last or current progress synchronization.",
	Example: "mangahub progress sync-status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)

		status, err := client.GetSyncStatus(cmd.Context())
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, status)
		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Progress Sync Status")
		cmd.Println("")
		cmd.Printf("Local database: ✓ Updated %s\n", api.HumanRelative(time.Now().UTC().Add(-status.LocalUpdatedAgo)))
		cmd.Printf("TCP sync server: ✓ %d devices connected\n", status.TCPDevices)
		cmd.Printf("Cloud backup: ✓ Last synced: %s\n", status.CloudLastSync.UTC().Format("2006-01-02 15:04:05 MST"))
		cmd.Println("")

		if status.CloudPendingDelta > 0 {
			cmd.Printf("⚠ Cloud backup behind by %d updates.\n", status.CloudPendingDelta)
			cmd.Println("Run: mangahub progress sync")
			return nil
		}

		cmd.Println("No issues detected.")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(syncStatusCmd)
}
