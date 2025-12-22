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
	Use:     "subscribe <novelID>",
	Short:   "Subscribe to notifications",
	Long:    "Subscribe to MangaHub notifications for a specific novel.",
	Example: "mangahub notify subscribe 42",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if strings.TrimSpace(cfg.Data.Token) == "" {
			return fmt.Errorf("✗ You must be logged in to manage notifications. Please login first.")
		}

		novelID, err := normalizeSubscriptionID(args[0])
		if err != nil {
			return err
		}

		client := notifyclient.NewUDPClient(cfg)

		output.PrintJSON(cmd, map[string]any{"delivery": "udp", "port": client.Port(), "novel_id": novelID})

		if !config.Runtime().Quiet {
			cmd.Println("Subscribing to chapter release notifications...")
		}

		if _, err := client.Subscribe(cmd.Context()); err != nil {
			return err
		}

		updated, err := addSubscription(cfg, novelID)
		if err != nil {
			return err
		}

		if config.Runtime().Quiet {
			if updated {
				cmd.Println(novelID)
			}
			return nil
		}

		if !updated {
			cmd.Println("Already subscribed to notifications for this novel.")
			return nil
		}

		cmd.Println("✓ Subscribed to novel notifications.")
		cmd.Println("You will now receive updates for this novel.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(subscribeCmd)
}
