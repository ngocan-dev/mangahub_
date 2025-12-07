package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:     "set <key> <value>",
	Short:   "Set a configuration value",
	Long:    "Set and persist a configuration key in the MangaHub config file.",
	Example: "mangahub config set server.host \"192.168.1.100\"\nmangahub config set notifications.enabled false",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		rawValue := args[1]

		mgr := config.ManagerInstance()
		if mgr == nil {
			return fmt.Errorf("✗ configuration not loaded")
		}

		expected, ok := validConfigKeys()[key]
		if !ok {
			suggestion := suggestKey(key)
			if suggestion != "" {
				return fmt.Errorf("✗ Invalid configuration key: %s\nDid you mean: %s ?", key, suggestion)
			}
			return fmt.Errorf("✗ Invalid configuration key: %s", key)
		}

		val, err := convertValue(rawValue, expected)
		if err != nil {
			return err
		}

		before, _ := json.MarshalIndent(mgr.Data, "", "  ")

		if err := applyKey(mgr, key, val); err != nil {
			return err
		}

		if err := mgr.Save(); err != nil {
			return err
		}

		runtime := config.Runtime()
		if runtime.Verbose {
			after, _ := json.MarshalIndent(mgr.Data, "", "  ")
			cmd.Println(string(before))
			cmd.Println("---")
			cmd.Println(string(after))
		}

		if runtime.Quiet {
			cmd.Println("✓ Configuration updated")
			return nil
		}

		cmd.Println("✓ Configuration updated")
		cmd.Printf("%s = %s\n", key, displayValue(val))
		cmd.Printf("Saved to: %s\n", humanizePath(mgr.Path))
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}

func validConfigKeys() map[string]string {
	return map[string]string{
		"server.host":           "string",
		"server.port":           "int",
		"server.grpc":           "int",
		"sync.tcp_port":         "int",
		"notify.udp_port":       "int",
		"chat.ws_port":          "int",
		"notifications.enabled": "bool",
		"notifications.sound":   "bool",
		"auth.username":         "string",
		"auth.profile":          "string",
	}
}

func convertValue(raw string, expected string) (interface{}, error) {
	raw = strings.Trim(raw, "\"")
	switch expected {
	case "bool":
		parsed, err := strconv.ParseBool(strings.ToLower(raw))
		if err != nil {
			return nil, fmt.Errorf("✗ %s is not a valid boolean", raw)
		}
		return parsed, nil
	case "int":
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("✗ %s is not a valid number", raw)
		}
		return parsed, nil
	default:
		return raw, nil
	}
}

func applyKey(mgr *config.Manager, key string, val interface{}) error {
	switch key {
	case "server.host":
		mgr.Data.Server.Host = val.(string)
	case "server.port":
		mgr.Data.Server.Port = val.(int)
	case "server.grpc":
		mgr.Data.Server.GRPC = val.(int)
	case "sync.tcp_port":
		mgr.Data.Sync.TCPPort = val.(int)
	case "notify.udp_port":
		mgr.Data.Notify.UDPPort = val.(int)
		mgr.Data.UDPPort = val.(int)
	case "chat.ws_port":
		mgr.Data.Chat.WSPort = val.(int)
	case "notifications.enabled":
		mgr.Data.Notifications.Enabled = val.(bool)
		mgr.Data.Settings.Notifications = val.(bool)
	case "notifications.sound":
		mgr.Data.Notifications.Sound = val.(bool)
	case "auth.username":
		mgr.Data.Auth.Username = val.(string)
	case "auth.profile":
		mgr.Data.Auth.Profile = val.(string)
	default:
		return fmt.Errorf("✗ Invalid configuration key: %s", key)
	}

	mgr.Data.BaseURL = fmt.Sprintf("http://%s:%d", mgr.Data.Server.Host, mgr.Data.Server.Port)
	mgr.Data.GRPCAddress = fmt.Sprintf("%s:%d", mgr.Data.Server.Host, mgr.Data.Server.GRPC)

	return nil
}

func displayValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	default:
		return fmt.Sprint(v)
	}
}

func suggestKey(invalid string) string {
	keys := make([]string, 0, len(validConfigKeys()))
	for k := range validConfigKeys() {
		keys = append(keys, k)
	}

	best := ""
	bestScore := -1
	for _, k := range keys {
		score := levenshtein(invalid, k)
		if bestScore == -1 || score < bestScore {
			best = k
			bestScore = score
		}
	}
	if bestScore <= 3 {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	aLen := len(a)
	bLen := len(b)
	dp := make([][]int, aLen+1)
	for i := 0; i <= aLen; i++ {
		dp[i] = make([]int, bLen+1)
		dp[i][0] = i
	}
	for j := 0; j <= bLen; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= aLen; i++ {
		for j := 1; j <= bLen; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}

	return dp[aLen][bLen]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

func humanizePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(path)
	}
	return strings.Replace(filepath.Clean(path), home, "~", 1)
}
