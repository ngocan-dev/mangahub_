
 # 📚 MangaHub – Network Programming Project (Go + Next.js)
+> A full-stack **Manga Management & Reading System** built for the **Net-Centric Programming** course.
+> This project demonstrates multi-protocol communication (HTTP, gRPC, TCP, UDP, WebSocket) using **Go** for the backend and CLI,
+> and **Next.js** for the frontend.
 
 ---
 
 ## 🚀 Overview
 
 **MangaHub** is a web and command-line platform that allows users to:
 - Browse and read manga online.
 - Leave comments, ratings, and track reading history.
 - Admins can add, edit, or delete manga titles.
 - Send **real-time notifications** to users via **UDP broadcast**.
 
 It’s designed to showcase network communication using multiple protocols and layered architecture.
 
 ---
+
+## 🧱 Monorepo Layout
+
+```
+mangahub_/
+├── backend/            # Go services (HTTP, gRPC, WebSocket, TCP, UDP)
+│   ├── cmd/            # Application entry points (binaries)
+│   ├── configs/        # YAML/JSON/TOML configuration files
+│   ├── internal/       # Clean-architecture layers (api, core, repository, service, transport)
+│   ├── migrations/     # SQL migrations for persistence
+│   ├── pkg/            # Reusable Go helpers (configuration, logger, etc.)
+│   └── proto/          # Protobuf contracts shared with clients
+├── cli/                # Go-powered CLI utility
+│   ├── cmd/            # CLI entry point (`main.go`)
+│   ├── internal/       # Commands, protocol clients, background sync tasks
+│   └── pkg/            # Shared utilities (config parsing, prompts)
+├── database/           # SQL schema, migrations, and seed data
+├── frontend/           # Next.js UI for browsing/reading manga
+└── README.md
+```
+
+Each major surface (backend API, CLI tooling, frontend UI) keeps its own README that explains coding conventions and folder-level responsibilities.
+
+---
+
+## 🛠️ Tech Stack
+
+| Layer     | Technology                                          |
+| --------- | --------------------------------------------------- |
+| Backend   | Go 1.21+, gRPC, WebSocket, TCP/UDP sockets          |
+| CLI       | Go, Cobra (planned), gRPC/UDP/TCP clients           |
+| Frontend  | Next.js 15, React, Tailwind CSS                     |
+| Database  | SQLite (development), open to PostgreSQL/MySQL swap |
+
+---
+
+## 👥 Team
+
 | Name                 | Role                    |
 | -------------------- | ----------------------- |
 | **HO NGOC AN**       | Developer, Backend, CLI |
 | **NGUYEN VIET THAO** | Developer, Frontend, UI |
 
 ---
 
+## ✅ Next Steps
+- Fill the scaffolded directories (`cmd`, `internal`, `pkg`, etc.) with implementation code.
+- Add CI workflows (linting, testing) once core services land.
+- Wire the frontend to the backend via HTTP/gRPC endpoints.
 
+Feel free to extend this layout with `docs/`, `deploy/`, or `infra/` folders as the project grows.
 