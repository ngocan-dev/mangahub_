package notify

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	notifyclient "github.com/ngocan-dev/mangahub_/cli/internal/notify"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// NotifyCmd manages notification preferences.
var NotifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Manage notifications",
	Long:  "Subscribe to, configure, and broadcast MangaHub notifications.",
}

var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to notifications",
	Long:  "Subscribe to MangaHub notifications for chapter releases and updates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if strings.TrimSpace(cfg.Data.Token) == "" {
			return fmt.Errorf("✗ You must be logged in to manage notifications. Please login first.")
		}

		client := notifyclient.NewUDPClient(cfg)

		output.PrintJSON(cmd, map[string]any{"delivery": "udp", "port": client.Port()})

		if !config.Runtime().Quiet {
			cmd.Println("Subscribing to chapter release notifications...")
		}

		updated, err := client.Subscribe(cmd.Context())
		if err != nil {
			return err
		}

		if config.Runtime().Quiet {
			if updated {
				cmd.Println("enabled")
			}
			return nil
		}

		if !updated {
			cmd.Println("Notifications are already enabled for this account on this device.")
			return nil
		}

		cmd.Println("✓ Notifications enabled for your account.")
		cmd.Println("You will now receive UDP alerts for new chapter releases on this device.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(subscribeCmd)
}
