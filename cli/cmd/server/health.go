package server

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:     "health",
	Short:   "Check server health",
	Long:    "Perform a health check against the MangaHub server.",
	Example: "mangahub server health",
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

		if config.Runtime().Quiet {
			return nil
		}

		printHealthSummary(cmd, status)
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(healthCmd)
	output.AddFlag(healthCmd)
}

func printHealthSummary(cmd *cobra.Command, status *api.ServerStatus) {
	if status == nil {
		return
	}

	cmd.Println("MangaHub Server Health")
	cmd.Println()

	cmd.Printf("Overall: %s\n", output.FormatOverall(status.Overall))
	cmd.Println()

	if len(status.Issues) > 0 {
		cmd.Println("Issues:")
		for _, issue := range status.Issues {
			cmd.Printf("  ✗ %s\n", issue)
		}
		cmd.Println()
	}

	if len(status.Services) > 0 {
		cmd.Println("Services:")
		for _, svc := range status.Services {
			cmd.Printf("  %s: %s (%s)\n", svc.Name, output.FormatStatus(svc.Status), svc.Load)
		}
		cmd.Println()
	}

	cmd.Println("Database:")
	if status.Database.Connection != "" {
		cmd.Printf("  Connection: %s\n", output.FormatConnection(status.Database.Connection))
	}
	if status.Database.Size != "" {
		cmd.Printf("  Size: %s\n", status.Database.Size)
	}
	if len(status.Database.Tables) > 0 {
		cmd.Printf("  Tables: %d\n", len(status.Database.Tables))
	}
	if status.Database.LastBackup != "" {
		cmd.Printf("  Last backup: %s\n", status.Database.LastBackup)
	}
	cmd.Println()

	if status.Resources.Memory != "" {
		cmd.Printf("Memory Usage: %s\n", status.Resources.Memory)
	}
	if status.Resources.CPU != "" {
		cmd.Printf("CPU Usage: %s\n", status.Resources.CPU)
	}
	if status.Resources.Disk != "" {
		cmd.Printf("Disk Space: %s\n", status.Resources.Disk)
	}
}
