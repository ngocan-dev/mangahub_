package profile

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

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
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("✗ Profile name is required.")
		}

		path, err := config.CreateProfile(name)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("✗ Profile '%s' already exists.", name)
			}
			return fmt.Errorf("✗ %v", err)
		}

		cmd.Printf("✓ Profile created: %s\n", name)
		cmd.Printf("Path: %s/\n", humanizePath(path))
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(createCmd)
	createCmd.Flags().String("name", "", "Profile name")
}
