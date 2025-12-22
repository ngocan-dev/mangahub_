package serve

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	serverstate "github.com/ngocan-dev/mangahub_/cli/internal/server"
	"github.com/spf13/cobra"
)

// ServeCmd starts or inspects the server stack by protocol.
var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Serve the MangaHub stack",
	Long:    "Inspect the MangaHub server stack and filter by protocol.",
	Example: "mangahub serve --protocol=http",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		protocol, _ := cmd.Flags().GetString("protocol")
		protocol = strings.ToLower(strings.TrimSpace(protocol))

		httpOnly, tcpOnly, udpOnly, grpcOnly, err := parseProtocol(protocol)
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

		status = serverstate.FilterByMode(status, httpOnly, tcpOnly, udpOnly, grpcOnly)

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
	ServeCmd.Flags().String("protocol", "", "Filter by protocol (tcp|udp|http|grpc)")
	output.AddFlag(ServeCmd)
}

func parseProtocol(protocol string) (bool, bool, bool, bool, error) {
	switch protocol {
	case "":
		return false, false, false, false, nil
	case "http":
		return true, false, false, false, nil
	case "tcp":
		return false, true, false, false, nil
	case "udp":
		return false, false, true, false, nil
	case "grpc":
		return false, false, false, true, nil
	default:
		return false, false, false, false, errors.New("--protocol must be one of: tcp, udp, http, grpc")
	}
}
