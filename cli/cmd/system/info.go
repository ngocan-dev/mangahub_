package system

import (
	"fmt"
	"runtime"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/version"
	"github.com/spf13/cobra"
)

// SystemCmd groups system-related helpers.
var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "System and diagnostic helpers",
	Long:  "System and diagnostic helpers for MangaHub CLI, including environment reports for bug submissions.",
}

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show MangaHub system information",
	Long:    "Print environment diagnostics useful for bug reports and support requests.",
	Example: "mangahub system info",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		configDir, _ := config.ConfigDir()
		details := map[string]any{
			"cli_version":       version.CLIVersion,
			"build_time":        version.BuildTime,
			"api_compatibility": version.APICompatibility,
			"active_profile":    cfg.ActiveProfile(),
			"config_path":       cfg.Path,
			"data_path":         configDir,
			"base_url":          cfg.Data.BaseURL,
			"grpc_address":      cfg.Data.GRPCAddress,
			"tcp_address":       cfg.Data.TCPAddress,
			"runtime": map[string]string{
				"os":         runtime.GOOS,
				"arch":       runtime.GOARCH,
				"go_version": runtime.Version(),
			},
			"server_components": map[string]any{
				"http_api":       fmt.Sprintf("%s:%d", cfg.Data.Server.Host, cfg.Data.Server.Port),
				"tcp_sync":       cfg.Data.TCPAddress,
				"udp_notify":     fmt.Sprintf("%s:%d", cfg.Data.Server.Host, cfg.Data.Notify.UDPPort),
				"grpc_service":   cfg.Data.GRPCAddress,
				"websocket_chat": fmt.Sprintf("%s:%d", cfg.Data.Server.Host, cfg.Data.Chat.WSPort),
			},
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, details)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, details)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("MangaHub System Information")
		cmd.Println()
		cmd.Printf("CLI Version:       %s\n", version.CLIVersion)
		cmd.Printf("Build:             %s\n", version.BuildTime)
		cmd.Printf("Backend API:       %s compatible\n", version.APICompatibility)
		cmd.Printf("Active Profile:    %s\n", cfg.ActiveProfile())
		cmd.Printf("Config Path:       %s\n", cfg.Path)
		cmd.Printf("Data Path:         %s\n", configDir)
		cmd.Printf("Backend API URL:   %s\n", cfg.Data.BaseURL)
		cmd.Printf("gRPC Address:      %s\n", cfg.Data.GRPCAddress)
		cmd.Printf("TCP Address:       %s\n", cfg.Data.TCPAddress)
		cmd.Println()
		cmd.Printf("OS:                %s/%s\n", runtime.GOOS, runtime.GOARCH)
		cmd.Printf("Go Runtime:        %s\n", runtime.Version())
		cmd.Println()
		cmd.Println("Server Components:")
		cmd.Printf("  HTTP API:        configured (%s:%d)\n", cfg.Data.Server.Host, cfg.Data.Server.Port)
		cmd.Printf("  TCP Sync:        configured (%s)\n", cfg.Data.TCPAddress)
		cmd.Printf("  UDP Notify:      configured (%s:%d)\n", cfg.Data.Server.Host, cfg.Data.Notify.UDPPort)
		cmd.Printf("  gRPC Service:    configured (%s)\n", cfg.Data.GRPCAddress)
		cmd.Printf("  WebSocket Chat:  configured (%s:%d)\n", cfg.Data.Server.Host, cfg.Data.Chat.WSPort)
		cmd.Println()
		cmd.Println("For bug reports, please include:")
		cmd.Println("  - The command you ran")
		cmd.Println("  - Output with --verbose")
		cmd.Println("  - Relevant logs from: mangahub logs errors")
		return nil
	},
}

func init() {
	SystemCmd.AddCommand(infoCmd)
	output.AddFlag(infoCmd)
}
