package server

import (
	"fmt"
	"strings"

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
		httpOnly, _ := cmd.Flags().GetBool("http-only")
		tcpOnly, _ := cmd.Flags().GetBool("tcp-only")
		udpOnly, _ := cmd.Flags().GetBool("udp-only")

		if err := validateStartFlags(httpOnly, tcpOnly, udpOnly); err != nil {
			return err
		}

		components := resolveComponents(httpOnly, tcpOnly, udpOnly)
		serverstate.MarkRunning(components)

		cmd.Println("Starting MangaHub Server Components...")
		cmd.Println()
		for idx, component := range components {
			cmd.Println(fmt.Sprintf("[%d/%d] %s", idx+1, len(components), component.Name))
			for _, line := range component.StartMessages {
				cmd.Println(line)
			}
			cmd.Println()
		}
		cmd.Println("All servers started successfully!")
		cmd.Println()
		cmd.Println("Server URLs:")
		cmd.Println("HTTP API:    http://localhost:8080")
		cmd.Println("TCP Sync:    tcp://localhost:9090")
		cmd.Println("UDP Notify:  udp://localhost:9091")
		cmd.Println("gRPC:        grpc://localhost:9092")
		cmd.Println("WebSocket:   ws://localhost:9093")
		cmd.Println()
		cmd.Println("Logs: tail -f ~/.mangahub/logs/server.log")
		cmd.Println("Stop: mangahub server stop")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(startCmd)
	startCmd.Flags().Bool("http-only", false, "Start only the HTTP server")
	startCmd.Flags().Bool("tcp-only", false, "Start only the TCP server")
	startCmd.Flags().Bool("udp-only", false, "Start only the UDP server")
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

func resolveComponents(httpOnly, tcpOnly, udpOnly bool) []serverstate.Component {
	components := serverstate.Components()
	if httpOnly {
		return []serverstate.Component{components[0]}
	}
	if tcpOnly {
		return []serverstate.Component{components[1]}
	}
	if udpOnly {
		return []serverstate.Component{components[2]}
	}
	return components
}
