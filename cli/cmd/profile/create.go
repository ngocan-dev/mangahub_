package profile

import "github.com/spf13/cobra"

// ProfileCmd manages CLI profiles.
var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage MangaHub profiles",
	Long:  "Create, switch, and list MangaHub CLI profiles.",
}

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a profile",
	Long:    "Create a new MangaHub CLI profile with its own configuration.",
	Example: "mangahub profile create --name work",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement profile creation
		cmd.Println("Profile creation is not yet implemented.")
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(createCmd)
	createCmd.Flags().String("name", "", "Profile name")
}
