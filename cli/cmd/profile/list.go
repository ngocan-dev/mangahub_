package profile

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profiles",
	Long:    "List all configured MangaHub CLI profiles.",
	Example: "mangahub profile list",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, active, err := config.ProfileList()
		if err != nil {
			return err
		}

		cmd.Println("Available Profiles:")
		cmd.Println()

		for _, p := range profiles {
			if p == "" {
				continue
			}
			if p == active {
				cmd.Printf("* %-8s (active)\n", p)
				continue
			}
			cmd.Printf("  %s\n", p)
		}
		return nil
	},
}

func init() {
	ProfileCmd.AddCommand(listCmd)
}
