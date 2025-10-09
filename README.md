# 📚 MangaHub – Network Programming Project (Go + Next.js)

> A full-stack **Manga Management & Reading System** built for the **Net-Centric Programming** course.  
> This project demonstrates multi-protocol communication (HTTP, gRPC, TCP, UDP, WebSocket) using **Go** for the backend and CLI, and **Next.js** for the frontend.

---

## 🚀 Overview

**MangaHub** is a web and command-line platform that allows users to:
- Browse and read manga online.
- Leave comments, ratings, and track reading history.
- Admins can add, edit, or delete manga titles.
- Send **real-time notifications** to users via **UDP broadcast**.

It’s designed to showcase network communication using multiple protocols and layered architecture.

---

## 🏗️ System Architecture

```

```
            ┌──────────────────────────┐
            │     Frontend (Next.js)    │
            │  - HTTP / WebSocket       │
            └────────────┬──────────────┘
                         │
          HTTP / REST / WS API
                         │
```

┌───────────────┐     ┌──────▼──────┐     ┌──────────────┐
│   CLI (Go)    │     │  Backend    │     │  Database     │
│ (Admin Tool)  │<--->│  (Go APIs)  │<--->│  SQLite / PG  │
│ TCP/gRPC/UDP  │     │             │     │               │
└───────────────┘     └──────┬──────┘     └──────────────┘
│
UDP Broadcast (Notify)
│
┌────▼────┐
│  Users  │
└─────────┘

```

---

## ⚙️ Features

| Role | Functionality | Protocol |
|------|----------------|-----------|
| **User** | View and read manga | HTTP / WebSocket |
|  | Comment and rate manga | HTTP |
|  | Receive notifications | UDP |
| **Admin** | Add / Update / Delete manga | gRPC |
|  | Sync CLI data | TCP |
|  | Broadcast new manga notifications | UDP |

---

## 🧩 Project Structure

```

mangahub/
├── backend/                      # Go backend (multi-protocol)
│   ├── cmd/
│   │   ├── api-server/           # HTTP API server
│   │   ├── grpc-server/          # gRPC service
│   │   ├── tcp-server/           # TCP sync server
│   │   └── udp-server/           # UDP notification server
│   ├── internal/
│   │   ├── auth/                 # Authentication logic
│   │   ├── manga/                # Manga management
│   │   ├── user/                 # User management
│   │   ├── websocket/            # Realtime chapter reader
│   │   ├── tcp/                  # TCP handler
│   │   ├── udp/                  # UDP broadcaster
│   │   └── grpc/                 # gRPC implementation
│   ├── pkg/
│   │   ├── database/             # DB init & migration
│   │   ├── models/               # Structs (Manga, User, etc.)
│   │   ├── utils/                # Helper functions
│   │   └── notifier/             # Shared UDP utilities
│   ├── proto/                    # Protocol Buffers (for gRPC)
│   ├── data/                     # Example JSON data
│   └── docs/                     # Documentation
│
├── cli/                          # Command-line admin tool
│   ├── commands/                 # Add / Delete / List / Sync commands
│   └── pkg/                      # CLI helpers
│       ├── config/               # CLI config (ports, IP)
│       ├── network/              # TCP/gRPC clients
│       └── utils/                # Print helpers
│
├── frontend/                     # Next.js web app
│   ├── pages/                    # Routes (index, manga, chapter, etc.)
│   ├── components/               # UI components
│   ├── lib/                      # API functions
│   ├── public/                   # Static assets
│   └── styles/                   # CSS / Tailwind
│
├── database/
│   ├── schema.sql                # Database schema
│   └── seed.sql                  # Sample data
│
├── docker-compose.yml            # Optional dev setup
└── README.md

````

---

## 🛠️ Installation

### 🧱 1. Backend (Go)
```bash
cd backend
go mod tidy
go run ./cmd/api-server
````

* Server runs on `http://localhost:8080`
* SQLite DB auto-created: `mangahub.db`

### 🧰 2. CLI (Go)

```bash
cd cli
go run main.go list       # Example: list mangas
go run main.go add "Naruto" --author "Kishimoto"
```

### 🌐 3. Frontend (Next.js)

```bash
cd frontend
npm install
npm run dev
```

* Frontend runs on `http://localhost:3000`
* Automatically fetches from backend API (`http://localhost:8080/api/...`)

---

## 🧠 Database Schema Overview

| Table             | Purpose                                 |
| ----------------- | --------------------------------------- |
| `users`           | Accounts and roles                      |
| `manga`           | Manga info (title, author, genre, etc.) |
| `chapter`         | Chapters per manga                      |
| `comment`         | User comments                           |
| `rating`          | User ratings                            |
| `reading_history` | Track chapters read                     |
| `notifications`   | UDP / web notification logs             |

---

## 🔌 Supported Protocols

| Protocol      | Used For                    | Implemented In               |
| ------------- | --------------------------- | ---------------------------- |
| **HTTP**      | REST API (frontend)         | `api-server`                 |
| **gRPC**      | Admin manga CRUD            | `grpc-server`                |
| **TCP**       | Sync data manually          | `tcp-server`                 |
| **UDP**       | Admin-to-user notifications | `udp-server`                 |
| **WebSocket** | Realtime manga reader       | `websocket` internal package |

---

## 💻 Quick Demo (HTTP)

**Start the backend:**

```bash
go run ./cmd/api-server
```

**Test endpoints:**

```bash
curl http://localhost:8080/api/health
curl -X POST http://localhost:8080/api/manga \
  -H "Content-Type: application/json" \
  -d '{"title":"Naruto","author":"Kishimoto","genre":"Action"}'
curl http://localhost:8080/api/manga
```

---

## 🧠 Learning Goals

This project demonstrates:

* Multi-protocol communication (HTTP, TCP, UDP, gRPC, WebSocket)
* Concurrent server design in Go
* REST API + CLI hybrid application
* Database design and schema migration
* Integration with modern frontend (Next.js)

---

## 👥 Team

| Name                 | Role                    |
| -------------------- | ----------------------- |
| **HO NGOC AN**       | Developer, Backend, CLI |
| **NGUYEN VIET THAO** | Developer, Frontend, UI |

---

## 📄 License

MIT License © 2025 – MangaHub Project Team

```

---

### ✅ Notes
- You can rename `backend` → `server` or `mangahub-server` if you like shorter URLs.
- This README fits **GitHub’s markdown style** and looks clean with code blocks, tables, and diagrams.
- It works perfectly as your course submission doc as well (teacher will easily see protocols + team).

---

Would you like me to make a **shorter version (for README top section)** with badges (like Go / Next.js / SQLite icons)? It’s nice for GitHub profile aesthetics.
```
