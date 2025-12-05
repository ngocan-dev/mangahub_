package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize MangaHub CLI configuration",
	Run: func(cmd *cobra.Command, args []string) {

		home, _ := os.UserHomeDir()
		base := filepath.Join(home, ".mangahub")

		dirs := []string{
			base,
			filepath.Join(base, "logs"),
		}

		for _, d := range dirs {
			if _, err := os.Stat(d); os.IsNotExist(err) {
				os.MkdirAll(d, 0755)
			}
		}

		configPath := filepath.Join(base, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			os.WriteFile(configPath, []byte(defaultConfig), 0644)
		}

		dbPath := filepath.Join(base, "data.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			os.WriteFile(dbPath, []byte{}, 0644) // empty file, backend can init later
		}

		fmt.Println("MangaHub CLI initialized successfully!")
	},
}

const defaultConfig = `
server:
  host: "localhost"
  http_port: 8080
  tcp_port: 9090
  udp_port: 9091
  grpc_port: 9092
  websocket_port: 9093

database:
  path: "~/.mangahub/data.db"

user:
  username: ""
  token: ""

sync:
  auto_sync: true
  conflict_resolution: "last_write_wins"

notifications:
  enabled: true
  sound: false

logging:
  level: "info"
  path: "~/.mangahub/logs/"
`

func init() {
	rootCmd.AddCommand(initCmd)
}
