
# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go workspace files
COPY go.work go.work.sum* ./

# Copy backend module
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /build/backend

# Download dependencies
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build argument to specify which server to build
ARG SERVER=api-server
ARG GOOS=linux
ARG GOARCH=amd64

# Build the binary with optimizations
RUN CGO_ENABLED=1 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -ldflags="-w -s" \
    -o /app/server \
    ./cmd/${SERVER}/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata

# Create non-root user
RUN addgroup -g 1000 mangahub && \
    adduser -D -u 1000 -G mangahub mangahub

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Create data directory for SQLite
RUN mkdir -p /app/data && \
    chown -R mangahub:mangahub /app

# Switch to non-root user
USER mangahub

# Expose necessary ports
EXPOSE 8080 50051 8081 9000 9001/udp

# Health check (adjust based on server type)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "--health-check"] || exit 1

# Run the server
ENTRYPOINT ["/app/server"]
