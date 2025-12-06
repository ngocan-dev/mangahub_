package notify

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	notifyclient "github.com/ngocan-dev/mangahub_/cli/internal/notify"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var unsubscribeCmd = &cobra.Command{
	Use:   "unsubscribe",
	Short: "Unsubscribe from notifications",
	Long:  "Stop receiving MangaHub notifications on this device.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := notifyclient.NewUDPClient(cfg)

		output.PrintJSON(cmd, map[string]any{"delivery": "udp", "port": client.Port()})

		if !config.Runtime().Quiet {
			cmd.Println("Unsubscribing from chapter release notifications...")
		}

		updated, err := client.Unsubscribe(cmd.Context())
		if err != nil {
			return err
		}

		if config.Runtime().Quiet {
			if updated {
				cmd.Println("disabled")
			}
			return nil
		}

		if !updated {
			cmd.Println("Notifications are already disabled for this account on this device.")
			return nil
		}

		cmd.Println("✓ Notifications disabled.")
		cmd.Println("You will no longer receive UDP alerts for new chapter releases on this device.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(unsubscribeCmd)
}
