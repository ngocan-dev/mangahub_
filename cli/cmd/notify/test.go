package notify

import (
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	notifyclient "github.com/ngocan-dev/mangahub_/cli/internal/notify"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test notification",
	Long:  "Trigger a test notification to verify UDP delivery.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := notifyclient.NewUDPClient(cfg)

		payload := "[TEST] New chapter available for: One Piece (chapter 1096)"

		output.PrintJSON(cmd, map[string]any{"delivery": "udp", "port": client.Port(), "message": payload})

		if !config.Runtime().Quiet {
			cmd.Println("Sending test notification...")
			cmd.Printf("UDP listener: waiting on port %d...\n", client.Port())
		}

		client.TriggerTestNotification(cmd.Context(), payload)
		msg, err := client.WaitForTestNotification(cmd.Context(), 3*time.Second)
		if err != nil {
			if err == notifyclient.ErrTimeout {
				return fmt.Errorf("✗ Test notification timed out. Check firewall or UDP settings.")
			}
			return err
		}

		if config.Runtime().Quiet {
			cmd.Println("ok")
			return nil
		}

		cmd.Println("✓ Test notification received:")
		cmd.Println(msg)
		cmd.Println("")
		cmd.Println("If you did not see any system notification, check your firewall or UDP settings.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(testCmd)
}
