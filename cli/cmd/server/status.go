package server

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show server status",
	Long:    "Display the current status of the MangaHub server.",
	Example: "mangahub server status",
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
		status, err := client.GetServerStatus(cmd.Context())
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

		output.PrintServerStatusTable(cmd, status)
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(statusCmd)
	output.AddFlag(statusCmd)
}
