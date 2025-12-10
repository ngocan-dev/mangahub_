package server

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the server",
	Long:    "Stop the running MangaHub server instance.",
	Example: "mangahub server stop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		status, err := client.GetServerStatus(cmd.Context())
		if err != nil {
			return err
		}

		serverstate.UpdateFromStatus(status)

		cmd.Println("Stopping MangaHub Servers...")
		cmd.Println()

		for _, component := range serverstate.Components() {
			cmd.Println("✓ " + component.StopLabel)
		}

		cmd.Println()
		cmd.Println("All services stopped.")
		serverstate.MarkStopped()
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(stopCmd)
}
