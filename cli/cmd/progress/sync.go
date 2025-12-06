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

		manual := isync.RunManualSync(resp.TCP.Devices, time.Now().UTC(), resp.Cloud.Pending)
		localMsg := "Updated"
		if resp.Local.Message != "" {
			localMsg = resp.Local.Message
		}
		cmd.Printf("Local database: %s %s\n", icon(resp.Local.OK), localMsg)

		tcpIcon := icon(resp.TCP.OK)
		tcpMsg := resp.TCP.Message
		if tcpMsg == "" {
			tcpMsg = fmt.Sprintf("Broadcasting latest progress to %d devices", manual.TCP.Devices)
		}
		if resp.TCP.OK && manual.TCP.Error != nil {
			tcpIcon = "✗"
			tcpMsg = manual.TCP.Error.Error()
		}
		cmd.Printf("TCP sync server: %s %s\n", tcpIcon, tcpMsg)

		cloudIcon := icon(resp.Cloud.OK)
		cloudMsg := resp.Cloud.Message
		if cloudMsg == "" {
			cloudMsg = manual.Cloud.Message
		}
		cmd.Printf("Cloud backup: %s %s\n", cloudIcon, cloudMsg)

		cmd.Println("")
		cmd.Println("✓ Sync completed successfully!")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(syncCmd)
}
