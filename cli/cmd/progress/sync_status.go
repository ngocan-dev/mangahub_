package progress

import (
	"fmt"

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
		cmd.Printf("Local database: %s Updated %s\n", icon(status.Local.OK), api.HumanRelative(status.Local.Updated))
		tcpMsg := status.TCP.Message
		if tcpMsg == "" {
			tcpMsg = fmt.Sprintf("%d devices connected", status.TCP.Devices)
		}
		cmd.Printf("TCP sync server: %s %s\n", icon(status.TCP.OK), tcpMsg)
		cmd.Printf("Cloud backup: %s Last synced: %s\n", icon(status.Cloud.OK), status.Cloud.LastSync.UTC().Format("2006-01-02 15:04:05 MST"))
		cmd.Println("")

		if status.Cloud.Pending > 0 {
			cmd.Printf("⚠ Cloud backup is behind the local database.\n")
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
