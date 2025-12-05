package config

// Config represents the MangaHub CLI configuration.
type Config struct {
	APIEndpoint   string `yaml:"api_endpoint"`
	GRPCAddress   string `yaml:"grpc_address"`
	DatabasePath  string `yaml:"database_path"`
	LogLevel      string `yaml:"log_level"`
	AuthToken     string `yaml:"auth_token"`
	ActiveProfile string `yaml:"active_profile"`
}

// Current holds the loaded configuration.
var Current Config

// Path tracks the loaded configuration file path.
var Path string
