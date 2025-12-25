package config

import (
	"fmt"
	"strings"
)

// ResolveBaseURL returns the HTTP API base URL after applying defaults.
func ResolveBaseURL(cfg Config) string {
	if trimmed := strings.TrimSpace(cfg.BaseURL); trimmed != "" {
		return strings.TrimRight(trimmed, "/")
	}
	host := cfg.Server.Host
	if host == "" {
		host = DefaultServerHost
	}
	port := cfg.Server.Port
	if port == 0 {
		port = DefaultServerPort
	}
	return fmt.Sprintf("%s://%s:%d", DefaultHTTPProtocol, host, port)
}

// ResolveGRPCAddress returns the gRPC endpoint with defaults applied.
func ResolveGRPCAddress(cfg Config) string {
	if trimmed := strings.TrimSpace(cfg.GRPCAddress); trimmed != "" {
		return trimmed
	}
	host := cfg.Server.Host
	if host == "" {
		host = DefaultServerHost
	}
	port := cfg.Server.GRPC
	if port == 0 {
		port = DefaultGRPCPort
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// ResolveTCPAddress returns the TCP sync address with defaults applied.
func ResolveTCPAddress(cfg Config) string {
	if trimmed := strings.TrimSpace(cfg.TCPAddress); trimmed != "" {
		return trimmed
	}
	host := cfg.Server.Host
	if host == "" {
		host = DefaultServerHost
	}
	port := cfg.Sync.TCPPort
	if port == 0 {
		port = DefaultSyncTCPPort
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// ResolveChatHost returns the chat host:port pair for both WS and HTTP history.
func ResolveChatHost(cfg Config) string {
	host := cfg.Server.Host
	if host == "" {
		host = DefaultServerHost
	}
	port := cfg.Chat.WSPort
	if port == 0 {
		port = DefaultChatWSPort
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// ResolveChatHTTPBase returns the HTTP base URL for chat history endpoints.
func ResolveChatHTTPBase(cfg Config) string {
	return fmt.Sprintf("%s://%s", DefaultHTTPProtocol, ResolveChatHost(cfg))
}

// ResolveChatWSURL builds the WebSocket URL for chat connections.
func ResolveChatWSURL(cfg Config, room string) string {
	host := ResolveChatHost(cfg)
	base := fmt.Sprintf("ws://%s/chat", host)
	if strings.TrimSpace(room) == "" {
		return base
	}
	return fmt.Sprintf("%s?room=%s", base, room)
}
