package config

// Config contains all application configuration grouped by subsystem.
type Config struct {
	App  AppConfig
	DB   DBConfig
	GRPC GRPCConfig
	UDP  UDPConfig
	Auth AuthConfig
}

type AppConfig struct {
	RedisURL      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	TCPServerAddr string
	WSServerAddr  string
}

type DBConfig struct {
	Driver        string
	DSN           string
	MigrationsDir string
	DatabaseURL   string
}

type GRPCConfig struct {
	ServerAddr string
}

type UDPConfig struct {
	ServerAddr        string
	MaxClients        int
	MaxClientsFromEnv bool
	Disabled          bool
}

type AuthConfig struct {
	JWTSecret string
}
