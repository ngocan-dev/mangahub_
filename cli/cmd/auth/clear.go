package auth

import "github.com/spf13/cobra"

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Clear stored authentication data",
	Long:    "Remove cached authentication credentials from the local environment.",
	Example: "mangahub auth clear",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement credential clearing
		cmd.Println("Authentication data clearing is not yet implemented.")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(clearCmd)
}
