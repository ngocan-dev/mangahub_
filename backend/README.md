# MangaHub Backend Layout

This Go module hosts every network-facing service that powers the platform. The goal is to keep concerns separated so that REST, gRPC, WebSocket, TCP, and UDP components stay maintainable as the codebase grows.

## Directory Guide

```
backend/
├── cmd/
│   └── server/           # Entry points for binaries (HTTP/gRPC gateway, workers, etc.)
├── configs/              # Static configuration files (YAML/JSON/TOML)
├── internal/
│   ├── api/
│   │   └── http/         # HTTP handlers + routing definitions
│   ├── core/             # Domain models and business logic contracts
│   ├── repository/       # Database gateways (SQL, cache, blob storage)
│   ├── service/          # Use-cases orchestrating core + repository layers
│   └── transport/
│       ├── grpc/         # gRPC services and protobuf-generated code wrappers
│       └── http/         # Middleware, server bootstrap, response utilities
├── migrations/           # SQL migrations and seeding data
├── pkg/                  # Reusable packages safe to import from other modules
│   └── config/           # Configuration loader helpers
└── proto/                # .proto definitions for shared contracts
```

### Coding Principles
- **Clean architecture** boundaries: handlers → services → repositories.
- **Protocol-first**: define protobuf/HTTP contracts before implementation.
- **Stateless services** where possible with configuration-driven behavior.

## Getting Started
1. Place your application entry point under `cmd/server` (e.g. `main.go`).
2. Implement domain logic inside `internal/core` and `internal/service`.
3. Generate protobuf output into `proto/` and expose via transports.
4. Add database migrations inside `migrations/`.