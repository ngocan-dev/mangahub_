package auth

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show authentication status",
	Long:    "Display the current authentication status and active profile.",
	Example: "mangahub auth status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement auth status check
		cmd.Println("Authentication status check is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(statusCmd)
}
