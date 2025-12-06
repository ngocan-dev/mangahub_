package config

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:     "reset",
	Short:   "Reset configuration",
	Long:    "Reset the MangaHub configuration file to default values.",
	Example: "mangahub config reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := config.ManagerInstance()
		if mgr == nil {
			return fmt.Errorf("✗ configuration not loaded")
		}

		cmd.Println("Resetting configuration to defaults...")

		if err := mgr.Reset(); err != nil {
			return err
		}

		runtime := config.Runtime()
		if runtime.Quiet {
			cmd.Println("✓ Configuration reset")
			cmd.Printf("Saved to: %s\n", humanizePath(mgr.Path))
			return nil
		}

		cmd.Println()
		cmd.Println("✓ Configuration reset")
		cmd.Printf("Saved to: %s\n", humanizePath(mgr.Path))
		cmd.Printf("Active profile: %s\n", mgr.ActiveProfile())
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(resetCmd)
}
