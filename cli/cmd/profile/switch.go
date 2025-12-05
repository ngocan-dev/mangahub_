package profile

import "github.com/spf13/cobra"

var switchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switch profile",
	Long:    "Switch the active MangaHub CLI profile.",
	Example: "mangahub profile switch --name work",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement profile switching
		cmd.Println("Profile switching is not yet implemented.")
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(switchCmd)
	switchCmd.Flags().String("name", "", "Profile name to activate")
}
