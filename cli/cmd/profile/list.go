package profile

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profiles",
	Long:    "List all configured MangaHub CLI profiles.",
	Example: "mangahub profile list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement profile listing
		cmd.Println("Profile listing is not yet implemented.")
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(listCmd)
}
