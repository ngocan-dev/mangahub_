package notify

import "github.com/spf13/cobra"

var unsubscribeCmd = &cobra.Command{
	Use:     "unsubscribe",
	Short:   "Unsubscribe from notifications",
	Long:    "Stop receiving MangaHub notifications for selected channels.",
	Example: "mangahub notify unsubscribe --channel email",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement notification unsubscribe
		cmd.Println("Notification unsubscribe is not yet implemented.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(unsubscribeCmd)
	unsubscribeCmd.Flags().String("channel", "email", "Notification channel")
}
