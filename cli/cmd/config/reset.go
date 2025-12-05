package config

import (
	"os"

	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:     "reset",
	Short:   "Reset configuration",
	Long:    "Reset the MangaHub configuration file to default values.",
	Example: "mangahub config reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.WriteFile(config.Path, []byte(config.DefaultConfigYAML), 0o644); err != nil {
			return err
		}
		cmd.Println("Configuration reset to defaults.")
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(resetCmd)
}
