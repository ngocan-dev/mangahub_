package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host          string `yaml:"host"`
		HTTPPort      int    `yaml:"http_port"`
		TCPPort       int    `yaml:"tcp_port"`
		UDPPort       int    `yaml:"udp_port"`
		GRPCPort      int    `yaml:"grpc_port"`
		WebsocketPort int    `yaml:"websocket_port"`
	} `yaml:"server"`

	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`

	User struct {
		Username string `yaml:"username"`
		Token    string `yaml:"token"`
	} `yaml:"user"`
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".mangahub", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	yaml.Unmarshal(data, cfg)
	return cfg, nil
}
