package config

import (
	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
)

// ConfigCmd manages CLI configuration files.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MangaHub CLI configuration",
	Long:  "Show and modify MangaHub CLI configuration settings.",
}

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show configuration",
	Long:    "Display the currently loaded MangaHub configuration.",
	Example: "mangahub config show",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Printf("Loaded config from %s\n", config.Path)
		cmd.Printf("API Endpoint: %s\n", config.Current.APIEndpoint)
		cmd.Printf("gRPC Address: %s\n", config.Current.GRPCAddress)
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}
