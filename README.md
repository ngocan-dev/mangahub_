# MangaHub

A production-ready manga reading platform demonstrating modern backend architecture with multi-protocol server design.

## Overview

MangaHub implements **HTTP REST, gRPC, WebSocket, TCP, and UDP servers**—each optimized for specific use cases (REST for integration, gRPC for performance, WebSocket for real-time chat, TCP for reliable sync, UDP for low-latency notifications).

**Key Features:**
- User management with JWT auth and RBAC
- Manga catalog with chapters, tags, and full-text search
- Reading history and cross-device sync
- Social features: ratings, comments, favorites, friends
- Real-time chat and notifications
- CLI tools for administration

**Tech Stack:** Go, Gin, gRPC, WebSocket, PostgreSQL/SQLite, Redis

## Architecture Highlights

- Clean Architecture with Domain-Driven Design
- Repository and Service patterns
- Redis caching and database indexing
- Security: JWT, bcrypt, rate limiting

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/ngocan-dev/mangahub.git
cd mangahub

# Start all services with Docker Compose
docker-compose up

# Or run individual servers
go run backend/cmd/api-server/main.go
go run backend/cmd/grpc-server/main.go
go run backend/cmd/ws-server/main.go
```

## 📧 Contact

For questions or collaboration:
- Developer: [Your Name]
- Email: [Your Email]
- GitHub: https://github.com/ngocan-dev/mangahub

---

**Note:** This is an educational/portfolio project demonstrating advanced backend development practices. It is not intended for commercial manga distribution.

