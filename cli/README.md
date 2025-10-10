# MangaHub CLI Layout

The CLI provides power-user access to the MangaHub ecosystem via TCP, UDP, and gRPC commands. This structure keeps the tooling aligned with the backend contracts while remaining easy to extend.

## Directory Guide

```
cli/
├── cmd/
│   └── mangahub/         # `main.go` entry point for the CLI binary
├── internal/
│   ├── commands/         # High-level command implementations (e.g. `login`, `sync`)
│   ├── sync/             # Background tasks, notification listeners, schedulers
│   └── transport/
│       ├── grpc/         # gRPC client stubs and wrappers
│       ├── tcp/          # Custom TCP client helpers
│       └── udp/          # UDP broadcaster/listener helpers
└── pkg/
    └── config/           # Shared configuration utilities (flags, env parsing)
```

### Development Notes
- Keep network clients reusable and protocol-agnostic inside `internal/transport`.
- Organize top-level commands by user intent inside `internal/commands`.
- Place reusable helpers in `pkg/` for future sharing with other modules.