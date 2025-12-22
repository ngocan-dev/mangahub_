package config

import "fmt"

// Default configuration constants.
const (
	DefaultBaseURL     = "http://localhost:8080"
	DefaultServerHost  = "localhost"
	DefaultServerPort  = 8080
	DefaultGRPCPort    = 9092
	DefaultSyncTCPPort = 9090
	DefaultNotifyUDP   = 9091
	DefaultChatWSPort  = 9093
)

// DefaultGRPCAddress builds the default gRPC address from host and port.
var DefaultGRPCAddress = fmt.Sprintf("%s:%d", DefaultServerHost, DefaultGRPCPort)

// DefaultUDPPort provides the fallback UDP port for notifications.
const DefaultUDPPort = DefaultNotifyUDP

// ServerConfig captures HTTP and gRPC ports.
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	GRPC int    `json:"grpc"`
}

// SyncConfig holds sync-specific options.
type SyncConfig struct {
	TCPPort int `json:"tcp_port"`
}

// NotifyConfig holds UDP notification options.
type NotifyConfig struct {
	UDPPort int `json:"udp_port"`
}

// ChatConfig stores websocket settings.
type ChatConfig struct {
	WSPort int `json:"ws_port"`
}

// NotificationsConfig contains user notification preferences.
type NotificationsConfig struct {
	Enabled       bool     `json:"enabled"`
	Sound         bool     `json:"sound"`
	Subscriptions []string `json:"subscriptions"`
}

// AuthSection captures authentication-related config.
type AuthSection struct {
	Username string `json:"username"`
	Profile  string `json:"profile"`
}

// LegacySettings preserves historical flags used elsewhere in the CLI.
type LegacySettings struct {
	Autosync      bool `json:"autosync"`
	Notifications bool `json:"notifications"`
}

// Config holds user-specific settings persisted to disk.
type Config struct {
	Server        ServerConfig        `json:"server"`
	Sync          SyncConfig          `json:"sync"`
	Notify        NotifyConfig        `json:"notify"`
	Chat          ChatConfig          `json:"chat"`
	Notifications NotificationsConfig `json:"notifications"`
	Auth          AuthSection         `json:"auth"`

	Token       string         `json:"token"`
	ExpiresAt   string         `json:"expires_at"`
	Permissions []string       `json:"permissions"`
	Settings    LegacySettings `json:"settings"`

	BaseURL     string `json:"base_url"`
	GRPCAddress string `json:"grpc_address"`
	UDPPort     int    `json:"udp_port"`
}

// DefaultConfig builds the baseline configuration for a given profile.
func DefaultConfig(profile string) Config {
	if profile == "" {
		profile = "default"
	}
	cfg := Config{
		Server: ServerConfig{
			Host: DefaultServerHost,
			Port: DefaultServerPort,
			GRPC: DefaultGRPCPort,
		},
		Sync:          SyncConfig{TCPPort: DefaultSyncTCPPort},
		Notify:        NotifyConfig{UDPPort: DefaultNotifyUDP},
		Chat:          ChatConfig{WSPort: DefaultChatWSPort},
		Notifications: NotificationsConfig{Enabled: true, Sound: true, Subscriptions: []string{}},
		Auth:          AuthSection{Username: "johndoe", Profile: profile},
		Settings:      LegacySettings{Notifications: true},
	}
	applyDerived(&cfg)
	return cfg
}

// applyDerived ensures compatibility fields stay in sync with defaults.
func applyDerived(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = DefaultServerHost
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = DefaultServerPort
	}
	if cfg.Server.GRPC == 0 {
		cfg.Server.GRPC = DefaultGRPCPort
	}
	if cfg.Sync.TCPPort == 0 {
		cfg.Sync.TCPPort = DefaultSyncTCPPort
	}
	if cfg.Notify.UDPPort == 0 {
		cfg.Notify.UDPPort = DefaultNotifyUDP
	}
	if cfg.Chat.WSPort == 0 {
		cfg.Chat.WSPort = DefaultChatWSPort
	}

	if !cfg.Notifications.Enabled && cfg.Settings.Notifications {
		cfg.Notifications.Enabled = true
	}
	if cfg.Notifications.Enabled {
		cfg.Settings.Notifications = true
	}

	if cfg.Auth.Profile == "" {
		cfg.Auth.Profile = "default"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	}
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPC)
	}
	if cfg.UDPPort == 0 {
		cfg.UDPPort = cfg.Notify.UDPPort
	}
}
