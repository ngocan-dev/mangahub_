# MangaHub CLI configuration and backend connectivity

The CLI now resolves backend addresses in a predictable order so it can talk to the running backend services out of the box.

## Precedence
`flags > environment variables > config file > defaults`

- Flags: `--server`, `--grpc`, `--tcp`
- Environment variables: `MANGAHUB_API`, `MANGAHUB_GRPC`, `MANGAHUB_TCP`
- Config file: `~/.mangahub/config.json` (created automatically)
- Defaults: `http://localhost:8080` (HTTP), `localhost:50051` (gRPC), `localhost:9000` (TCP), UDP notify `9091`, chat `9093`

Authentication tokens and resolved addresses are stored in `~/.mangahub/config.json` so subsequent runs reuse them automatically.

## Flag parsing and config loading
```go
// cmd/root.go
rootCmd.PersistentFlags().StringVar(&apiOverride,  "server", "", "HTTP API base URL (default: http://localhost:8080)")
rootCmd.PersistentFlags().StringVar(&grpcOverride, "grpc",   "", "gRPC server address (default: localhost:50051)")
rootCmd.PersistentFlags().StringVar(&tcpOverride,  "tcp",    "", "TCP server address (default: localhost:9000)")

config.LoadWithOptions(config.LoadOptions{
    Path:        cfgFile,       // falls back to ~/.mangahub/config.json
    APIEndpoint: apiOverride,   // flag > env > config > defaults
    GRPCAddress: grpcOverride,
    TCPAddress:  tcpOverride,
})
```

## Making an authenticated request
```go
// cmd/server/status.go
cfg := config.ManagerInstance()
client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token) // token persisted after login
status, err := client.GetServerStatus(cmd.Context())
```
The `api.Client` automatically attaches `Authorization: Bearer <token>` when a token is present in the config file.

## Running the CLI with the backend
- `go run ./cmd/mangahub status`
- `go run ./cmd/mangahub --server http://localhost:8080 status`
- `MANGAHUB_API=http://localhost:8080 go run ./cmd/mangahub status`

All commands will respect the same resolution order and reuse the stored token when calling the backend.
