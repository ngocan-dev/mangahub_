package notify

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List notification subscriptions",
	Long:    "List the novels you are subscribed to for notifications.",
	Example: "mangahub notify list",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		subscriptions := listSubscriptions(cfg)
		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"subscriptions": subscriptions})
			return nil
		}

		if config.Runtime().Quiet {
			for _, novelID := range subscriptions {
				cmd.Println(novelID)
			}
			return nil
		}

		if len(subscriptions) == 0 {
			cmd.Println("No notification subscriptions yet.")
			cmd.Println("Use: mangahub notify subscribe <novelID>")
			return nil
		}

		table := utils.Table{Headers: []string{"Novel ID"}}
		for _, novelID := range subscriptions {
			table.AddRow(novelID)
		}

		cmd.Println(table.RenderWithTitle("Notification Subscriptions"))
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(listCmd)
	output.AddFlag(listCmd)
}
