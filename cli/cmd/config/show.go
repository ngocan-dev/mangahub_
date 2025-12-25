package config

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

// ConfigCmd manages CLI configuration files.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MangaHub CLI configuration",
	Long:  "Show, modify, and reset MangaHub CLI configuration settings.",
}

var showCmd = &cobra.Command{
	Use:     "show [section]",
	Short:   "Show configuration",
	Long:    "Display the currently loaded MangaHub configuration. Provide an optional section to filter output.",
	Example: "mangahub config show\nmangahub config show server",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("✗ configuration not loaded")
		}

		if len(args) == 0 {
			printFullConfig(cmd, cfg.Data)
			return nil
		}

		section := args[0]
		if !printSection(cmd, section, cfg.Data) {
			return fmt.Errorf("✗ Section '%s' not found in configuration.", section)
		}
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}

func printFullConfig(cmd *cobra.Command, cfg config.Config) {
	cmd.Println("MangaHub Configuration")
	cmd.Println()

	cmd.Printf("%-20s %v\n", "api.base_url:", cfg.BaseURL)
	cmd.Printf("%-20s %v\n", "grpc.address:", cfg.GRPCAddress)
	cmd.Printf("%-20s %v\n", "tcp.address:", cfg.TCPAddress)
	cmd.Printf("%-20s %v\n", "server.host:", cfg.Server.Host)
	cmd.Printf("%-20s %v\n", "server.port:", cfg.Server.Port)
	cmd.Printf("%-20s %v\n", "server.grpc:", cfg.Server.GRPC)
	cmd.Printf("%-20s %v\n", "sync.tcp_port:", cfg.Sync.TCPPort)
	cmd.Printf("%-20s %v\n", "notify.udp_port:", cfg.Notify.UDPPort)
	cmd.Printf("%-20s %v\n", "chat.ws_port:", cfg.Chat.WSPort)

	cmd.Println()
	cmd.Printf("%-22s %v\n", "notifications.enabled:", cfg.Notifications.Enabled)
	cmd.Printf("%-22s %v\n", "notifications.sound:", cfg.Notifications.Sound)

	cmd.Println()
	cmd.Printf("%-15s %s\n", "auth.username:", cfg.Auth.Username)
	cmd.Printf("%-15s %s\n", "auth.profile:", cfg.Auth.Profile)
}

func printSection(cmd *cobra.Command, section string, cfg config.Config) bool {
	switch section {
	case "api":
		cmd.Println("[api]")
		cmd.Printf("base_url: %s\n", cfg.BaseURL)
	case "addresses", "endpoints":
		cmd.Println("[endpoints]")
		cmd.Printf("api:  %s\n", cfg.BaseURL)
		cmd.Printf("grpc: %s\n", cfg.GRPCAddress)
		cmd.Printf("tcp:  %s\n", cfg.TCPAddress)
	case "server":
		cmd.Println("[server]")
		cmd.Printf("host: %s\n", cfg.Server.Host)
		cmd.Printf("port: %d\n", cfg.Server.Port)
		cmd.Printf("grpc: %d\n", cfg.Server.GRPC)
	case "sync":
		cmd.Println("[sync]")
		cmd.Printf("tcp_port: %d\n", cfg.Sync.TCPPort)
	case "notify":
		cmd.Println("[notify]")
		cmd.Printf("udp_port: %d\n", cfg.Notify.UDPPort)
	case "chat":
		cmd.Println("[chat]")
		cmd.Printf("ws_port: %d\n", cfg.Chat.WSPort)
	case "notifications":
		cmd.Println("[notifications]")
		cmd.Printf("enabled: %v\n", cfg.Notifications.Enabled)
		cmd.Printf("sound: %v\n", cfg.Notifications.Sound)
	case "auth":
		cmd.Println("[auth]")
		cmd.Printf("username: %s\n", cfg.Auth.Username)
		cmd.Printf("profile: %s\n", cfg.Auth.Profile)
	default:
		return false
	}
	return true
}
