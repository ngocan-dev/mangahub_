package server

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

// ServerCmd controls the MangaHub server lifecycle.
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the MangaHub server",
	Long:  "Start, stop, and monitor the MangaHub server.",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server",
	Long:  "Start the MangaHub multi-protocol server stack.",
	Example: strings.Join([]string{
		"mangahub server start",
		"mangahub server start --http-only",
		"mangahub server start --tcp-only",
		"mangahub server start --udp-only",
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		httpOnly, _ := cmd.Flags().GetBool("http-only")
		tcpOnly, _ := cmd.Flags().GetBool("tcp-only")
		udpOnly, _ := cmd.Flags().GetBool("udp-only")

		if err := validateStartFlags(httpOnly, tcpOnly, udpOnly); err != nil {
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

		status = serverstate.FilterByMode(status, httpOnly, tcpOnly, udpOnly)

		if format == output.FormatJSON {
			output.PrintJSON(cmd, status)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, status)
		}

		cmd.Println("MangaHub Server Status")
		output.PrintServerStatusTable(cmd, status)
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(startCmd)
	startCmd.Flags().Bool("http-only", false, "Start only the HTTP server")
	startCmd.Flags().Bool("tcp-only", false, "Start only the TCP server")
	startCmd.Flags().Bool("udp-only", false, "Start only the UDP server")
	output.AddFlag(startCmd)
}

func validateStartFlags(httpOnly, tcpOnly, udpOnly bool) error {
	trueCount := 0
	for _, flag := range []bool{httpOnly, tcpOnly, udpOnly} {
		if flag {
			trueCount++
		}
	}
	if trueCount > 1 {
		return fmt.Errorf("only one of --http-only, --tcp-only, or --udp-only can be set")
	}
	return nil
}
