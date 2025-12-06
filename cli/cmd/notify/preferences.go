package notify

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	notifyclient "github.com/ngocan-dev/mangahub_/cli/internal/notify"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var preferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "View notification preferences",
	Long:  "Show notification delivery settings for this device.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := notifyclient.NewUDPClient(cfg)
		prefs := client.Preferences()

		output.PrintJSON(cmd, prefs)
		if config.Runtime().Quiet {
			if prefs.Enabled {
				cmd.Println("enabled")
			} else {
				cmd.Println("disabled")
			}
			return nil
		}

		cmd.Println("Notification Preferences")
		cmd.Println("")
		status := "Disabled"
		if prefs.Enabled {
			status = "Enabled"
		}
		cmd.Printf("Status: %s\n", status)
		cmd.Printf("Delivery: %s\n", prefs.Delivery)
		cmd.Printf("UDP Port: %d\n", prefs.Port)

		if prefs.Enabled {
			cmd.Println("Subscribed events:")
			for _, event := range prefs.Events {
				cmd.Printf("- %s\n", event)
			}
			cmd.Println("")
			cmd.Println("To disable notifications:")
			cmd.Println("mangahub notify unsubscribe")
			return nil
		}

		cmd.Println("No events will be sent to this device.")
		cmd.Println("")
		cmd.Println("To enable notifications:")
		cmd.Println("mangahub notify subscribe")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(preferencesCmd)
}
