# MangaHub Backend Architecture and Folder Layout

This document standardizes the MangaHub Go backend into a clean-architecture layout with multiple protocol servers, clear domain separation, and production-friendly operations.

## Canonical Folder Tree (with purposes)
```
mangahub/
  cmd/                                  Entry points for each binary (HTTP API, gRPC, TCP sync, UDP notify, WebSocket chat, workers).
    api-server/                         HTTP API server bootstrap (routing, middleware wiring, DI root for HTTP).
    grpc-server/                        gRPC server bootstrap (service registration, reflection, health checks).
    tcp-server/                         TCP sync server bootstrap (connection handling, framing, worker pool setup).
    udp-server/                         UDP notification server bootstrap (listener lifecycle, packet dispatcher wiring).
    websocket-server/                   WebSocket chat server bootstrap (hub initialization, upgrader config).
    migrator/                           Database migration runner entrypoint.

  internal/                             Application code not meant to be imported by external modules (clean architecture core).
    config/                             Config loader (env/files), validation, and typed settings structs.
    logger/                             Structured logging setup and adapters.
    security/                           AuthN/Z helpers, password hashing, JWT/session logic, CSRF/cors configs.
    middleware/                         HTTP and gRPC middleware (auth, request ID, logging, rate limiting, recovery).
    transport/                          Protocol-specific adapters.
      http/                             HTTP server wiring: routers, handlers, request/response DTOs, validators.
        handlers/                       HTTP handlers for each domain (manga, user, chapter, history, favorite, comment, rating).
        router/                         Route registration and middleware chains.
      grpc/                             gRPC server registration and interceptors.
        services/                       gRPC service implementations wrapping use cases.
      tcp/                              TCP server implementation (connection manager, framing, sync events dispatcher).
      udp/                              UDP server implementation (notification endpoints, packet serializer/deserializer).
      websocket/                        WebSocket hub (rooms, broadcast, presence) and handler for upgrades.
    domain/                             Enterprise business models and interfaces (pure go, no framework dependencies).
      manga/                            Manga aggregates/entities, value objects, and repository/service interfaces.
      user/                             User entities, authentication rules, and interfaces.
      chapter/                          Chapter entities and interfaces.
      history/                          Reading history entities and interfaces.
      favorite/                         Favorite entities and interfaces.
      comment/                          Comment entities and interfaces.
      rating/                           Rating entities and interfaces.
    repository/                         Data access implementations following domain interfaces.
      postgres/                         PostgreSQL repositories, queries, and transaction helpers.
      cache/                            Cache layer implementations (Redis/memory) adhering to cache interfaces.
      mock/                             In-memory fakes for testing.
    service/                            Application services/use cases orchestrating domain logic.
      manga/                            Manga service implementation (CRUD, search, tag filtering, recommendations).
      user/                             User service (registration, login, profiles, roles).
      chapter/                          Chapter service (CRUD, ordering, availability checks).
      history/                          History service (tracking progress, resume logic).
      favorite/                         Favorite service (collections, follow/unfollow logic).
      comment/                          Comment service (posting, moderation, threading).
      rating/                           Rating service (scores, aggregates, recommendations).
    queue/                              Asynchronous job processor definitions and workers.
    cache/                              Cache interfaces and helpers (TTL policies, key builders).
    auth/                               Authentication helpers shared across transports (token parsing, context helpers).
    events/                             Domain events, pub/sub contracts, and integration event handlers.

  pkg/                                  Reusable packages that could be open-sourced (pure utility, no app deps).
    httpx/                              HTTP utilities (response helpers, pagination, renderers).
    grpcx/                              gRPC utilities (error mapping, metadata helpers).
    netx/                               Network utilities (framing, retry dialers, backoff strategies).
    dbx/                                Database utilities (SQL helpers, migrations runner wrappers).
    loggerx/                            Logging helpers (slog/zap wrappers, testing logs).
    configx/                            Config utilities (env loader, watcher).

  proto/                                Protocol Buffer definitions and generated stubs for gRPC/WebSocket/TCP schemas.
    manga.proto                         Manga service definitions and messages.
    user.proto                          User service definitions and messages.
    chapter.proto                       Chapter service definitions and messages.
    history.proto                       History service definitions and messages.
    favorite.proto                      Favorite service definitions and messages.
    comment.proto                       Comment service definitions and messages.
    rating.proto                        Rating service definitions and messages.
    common.proto                        Shared messages (pagination, error, auth tokens).

  db/                                   Database artifacts and tooling.
    migrations/                         SQL migration files (timestamped up/down scripts).
    seeds/                              Seed data scripts for local/dev environments.
    fixtures/                           Test fixtures for repositories.

  data/                                 Static JSON/sample data for demos, tests, and grading assets.

  deployments/                          Deployment manifests (Docker/K8s), CI/CD scripts, and infrastructure as code.
    docker/                             Dockerfiles for each service/binary.
    k8s/                                Kubernetes manifests/Helm charts.
    terraform/                          Optional IaC modules for cloud resources.

  docs/                                 Documentation (architecture, ADRs, setup, runbooks).
    ADRs/                               Architectural decision records.

  scripts/                              Developer scripts (lint, test, generate, compose helpers).

  README.md                             Project overview and quickstart.
  docker-compose.yml                    Local development composition for all servers and dependencies.
```

## Component Placement Overview
- **HTTP handlers**: `internal/transport/http/handlers` with domain-specific subpackages and request/response DTOs.
- **gRPC services**: `internal/transport/grpc/services`, generated code in `proto/` with Go output to `internal/transport/grpc/pb` or `pkg/grpcx/pb` depending on visibility.
- **TCP server**: `internal/transport/tcp` (framing, sync protocol, worker pool, domain event bridge).
- **UDP server**: `internal/transport/udp` (notification fan-out, serializer, rate limiting).
- **WebSocket hub**: `internal/transport/websocket` (hub, rooms, presence, chat handlers mounted by HTTP router).
- **Domain models**: `internal/domain/<bounded-context>` packages.
- **Repository layer**: `internal/repository/<backend>` implementing `internal/domain/...` interfaces.
- **Service/use-case layer**: `internal/service/<context>` coordinating repositories, cache, and events.
- **Config loader**: `internal/config` using `pkg/configx` helpers.
- **Logger**: `internal/logger` with `pkg/loggerx` helpers.
- **Middleware**: `internal/middleware` for HTTP/gRPC; TCP/UDP hooks live in respective transport packages.
- **Queue processor**: `internal/queue` with worker registrations and background runners.
- **Cache interfaces/impl**: `internal/cache` for interfaces + helpers; `internal/repository/cache` for concrete Redis/memory cache stores.

## Migration Guidelines (moving existing code)
1. **Identify entrypoints**: Move existing HTTP/gRPC/TCP/UDP/WebSocket mains into `cmd/<server>/main.go`, keeping server-specific config flags and DI wiring there.
2. **Extract domain models**: Relocate plain structs and interfaces (manga, user, chapter, history, favorite, comment, rating) into `internal/domain/<context>`; strip transport/framework imports.
3. **Refactor services**: Move business logic currently in handlers/controllers into `internal/service/<context>`; inject interfaces instead of concrete DB clients.
4. **Create repositories**: For each domain, implement persistence in `internal/repository/postgres` (or other DB backends) satisfying the domain interfaces; move raw SQL/migrations to `db/migrations`.
5. **Transport adapters**: Rewrite handlers/controllers to depend only on service interfaces; place HTTP handlers under `internal/transport/http/handlers`, gRPC implementations under `internal/transport/grpc/services`, TCP/UDP servers under their respective folders, and WebSocket chat hub under `internal/transport/websocket`.
6. **Cross-cutting concerns**: Centralize config parsing in `internal/config`, logging in `internal/logger`, and security/auth helpers in `internal/security` with middleware in `internal/middleware`.
7. **Testing**: Use `internal/repository/mock` for in-memory fakes; keep unit tests alongside packages (e.g., `internal/service/manga/manga_test.go`).
8. **Utilities**: Move reusable helpers to `pkg/` so they remain import-clean; avoid importing `internal/` from `pkg/` packages to preserve boundaries.
9. **Data and fixtures**: Relocate JSON samples to `data/`; DB test fixtures belong in `db/fixtures`.
10. **Deployment assets**: Place Dockerfiles/K8s manifests in `deployments/` and consolidate compose setup into root `docker-compose.yml`.

## Dependency Injection Layout
- **Construction root per binary**: Each `cmd/<server>/main.go` initializes config (`internal/config`), logger (`internal/logger`), DB connections (`pkg/dbx`), cache (`internal/cache`), queue (`internal/queue`), and services (`internal/service/...`), then wires transport adapters.
- **Provider sets**: Use lightweight constructors (or wirefx/uber-go/fx if desired) grouped by layer: `config.Provider()`, `logger.New()`, `repository.NewPostgresX()`, `cache.NewRedisCache()`, `service.NewMangaService(repo, cache, events)`, `transport/http/handlers.New(...)`.
- **Interfaces first**: Transports depend on service interfaces; services depend on repository/cache interfaces; repositories depend on lower-level DB clients.
- **Shared kernels**: Security/auth middleware receive token validators from `internal/security`; event bus from `internal/events` can be injected where domain events are published/consumed.

## Bootstrap Expectations for `main.go`
- **api-server**: Parse config, init logger, DB, cache, queue; construct services; wire HTTP router + middleware (auth, rate limit, recovery); mount WebSocket handler; start server with graceful shutdown.
- **grpc-server**: Similar setup, register gRPC services from `internal/transport/grpc/services`, enable health/reflection, and start with graceful stop.
- **tcp-server**: Initialize config/logger, create TCP listener, setup framing/worker pool, inject services for sync operations, and handle reconnections with backoff.
- **udp-server**: Initialize config/logger, bind UDP socket, set notification dispatcher using services/cache, and run packet processing loop with metrics.
- **websocket-server**: Optionally standalone; initialize hub, auth middleware, register rooms/channels, and start HTTP server that upgrades connections.
- **migrator**: Load config, connect to DB, run `db/migrations` via `pkg/dbx` runner (up/down/redo) with logging.

## Bonus Recommendations for Grading Excellence
- Add CI to run `go test ./...`, `golangci-lint`, protobuf generation, and migration verification.
- Provide make/just scripts for `lint`, `test`, `generate`, and `compose-up/down`.
- Include sample `.env.example` and `docs/SETUP.md` covering local/dev/prod configs and TLS guidance.
- Document API (OpenAPI) and gRPC (buf/evans) usage in `docs/` and expose `/healthz` + `/readyz` endpoints.
- Add observability hooks (Prometheus metrics, structured logs, request IDs, tracing) with middleware in all transports.
- Supply load-testing profiles (k6) and chaos/resilience notes to demonstrate production-readiness.
