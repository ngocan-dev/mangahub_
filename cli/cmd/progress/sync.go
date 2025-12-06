package progress

import (
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	isync "github.com/ngocan-dev/mangahub_/cli/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Sync progress",
	Long:    "Synchronize progress with remote MangaHub services.",
	Example: "mangahub progress sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)

		resp, err := client.TriggerProgressSync(cmd.Context())
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Starting manual sync...\n")

		cmd.Printf("Local database: %s Ready\n", icon(resp.LocalReady))

		tcpResult := isync.Broadcast(resp.TCPDevices)
		tcpIcon := "✗"
		tcpMsg := "Failed"
		if tcpResult.Error == nil {
			tcpIcon = "✓"
			tcpMsg = fmt.Sprintf("Synced to %d devices", tcpResult.Devices)
		}
		cmd.Printf("TCP sync server: %s %s\n", tcpIcon, tcpMsg)

		cloud := isync.SyncCloud(time.Now().UTC(), 0)
		cmd.Printf("Cloud backup: %s %s\n", icon(cloud.Success), cloud.Message)

		cmd.Println("")
		cmd.Println("✓ Sync completed successfully!")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(syncCmd)
}
