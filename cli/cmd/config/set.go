package config

import (
	"fmt"
	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set a configuration value",
	Long:    "Set and persist a configuration key in the MangaHub config file.",
	Example: "mangahub config set api_endpoint https://api.example.com",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("key and value are required")
		}
		key := args[0]
		value := args[1]
		cmd.Printf("TODO: set %s to %s in %s\n", key, value, config.Path)
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
