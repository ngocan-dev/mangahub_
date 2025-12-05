package notify

import "github.com/spf13/cobra"

var preferencesCmd = &cobra.Command{
	Use:     "preferences",
	Short:   "Configure notification preferences",
	Long:    "Set notification channels, frequency, and filters for MangaHub notifications.",
	Example: "mangahub notify preferences --channel email --frequency daily",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement preference configuration
		cmd.Println("Notification preferences are not yet implemented.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(preferencesCmd)
	preferencesCmd.Flags().String("channel", "email", "Notification channel")
	preferencesCmd.Flags().String("frequency", "daily", "Notification frequency")
}
