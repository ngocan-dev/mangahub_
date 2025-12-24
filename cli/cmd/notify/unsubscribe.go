package notify

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	notifyclient "github.com/ngocan-dev/mangahub_/cli/internal/notify"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var unsubscribeCmd = &cobra.Command{
	Use:     "unsubscribe <novelID>",
	Short:   "Unsubscribe from notifications",
	Long:    "Stop receiving MangaHub notifications for a specific novel.",
	Example: "mangahub notify unsubscribe 42",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

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

		if format != output.FormatJSON && !config.Runtime().Quiet {
			cmd.Println("Unsubscribing from chapter release notifications...")
		}

		updated, err := removeSubscription(cfg, novelID)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{
				"delivery":     "udp",
				"port":         client.Port(),
				"novel_id":     novelID,
				"unsubscribed": updated,
			})
			return nil
		}

		if config.Runtime().Quiet {
			if updated {
				cmd.Println(novelID)
			}
			return nil
		}

		if !updated {
			cmd.Println("You are not subscribed to this novel.")
			return nil
		}

		cmd.Println("✓ Unsubscribed from novel notifications.")
		cmd.Println("You will no longer receive updates for this novel.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(unsubscribeCmd)
	output.AddFlag(unsubscribeCmd)
}
