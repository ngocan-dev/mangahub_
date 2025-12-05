package notify

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Send a test notification",
	Long:    "Trigger a test notification to verify configuration.",
	Example: "mangahub notify test --channel email",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement test notification
		cmd.Println("Notification test is not yet implemented.")
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(testCmd)
	testCmd.Flags().String("channel", "email", "Notification channel")
}
