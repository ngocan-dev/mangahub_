package server

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the server",
	Long:    "Stop the running MangaHub server instance.",
	Example: "mangahub server stop",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		status, err := serverstate.FetchStatus(cmd.Context(), client)
		if err != nil {
			return err
		}
		if format == output.FormatJSON {
			output.PrintJSON(cmd, status)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, status)
		}

		cmd.Println("Stopping MangaHub Servers...")
		cmd.Println()

		for _, svc := range status.Services {
			cmd.Printf("✓ %s at %s: %s (%s)\n", svc.Name, svc.Address, output.FormatStatus(svc.Status), svc.Load)
		}

		cmd.Println()
		cmd.Printf("Overall system state: %s\n", output.FormatOverall(status.Overall))
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(stopCmd)
	output.AddFlag(stopCmd)
}
