package profile

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switch profile",
	Long:    "Switch the active MangaHub CLI profile.",
	Example: "mangahub profile switch --name work",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("✗ Profile name is required.")
		}

		cmd.Printf("Switching profile to '%s'...\n\n", name)

		path, err := config.SwitchProfile(name)
		if err != nil {
			return fmt.Errorf("✗ Profile '%s' does not exist.\nUse: mangahub profile list", name)
		}

		if _, err := config.Load(""); err != nil {
			return err
		}

		cmd.Printf("✓ Active profile: %s\n", name)
		cmd.Println("Config loaded from:")
		cmd.Println(humanizePath(path))
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(switchCmd)
	switchCmd.Flags().String("name", "", "Profile name to activate")
}
