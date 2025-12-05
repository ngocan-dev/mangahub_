package config

const DefaultConfigYAML = `# MangaHub configuration
api_endpoint: "https://api.mangahub.local"
grpc_address: "localhost:50051"
database_path: "~/.mangahub/data.db"
log_level: "info"
auth_token: ""
active_profile: "default"
`

// DefaultConfig returns the default configuration structure.
func DefaultConfig() Config {
	return Config{
		APIEndpoint:   "https://api.mangahub.local",
		GRPCAddress:   "localhost:50051",
		DatabasePath:  "~/.mangahub/data.db",
		LogLevel:      "info",
		AuthToken:     "",
		ActiveProfile: "default",
	}
}
