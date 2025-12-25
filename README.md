# MangaHub - Hệ thống đọc truyện tranh đa giao thức

MangaHub là một ứng dụng web đọc truyện tranh (manga) được xây dựng với kiến trúc microservice, hỗ trợ nhiều giao thức giao tiếp khác nhau (HTTP, gRPC, TCP, UDP, WebSocket).

## 🏗️ Kiến trúc hệ thống

### Backend (Go)
- **HTTP API Server** - RESTful API cho ứng dụng web
- **gRPC Server** - Giao tiếp hiệu năng cao
- **TCP Server** - Đồng bộ dữ liệu
- **UDP Server** - Thông báo real-time
- **WebSocket Server** - Chat và cập nhật trực tiếp
- **Database**: SQLite (local development)

### Frontend (Next.js 14)
- React 18 + TypeScript
- TailwindCSS cho styling
- Axios cho HTTP requests

### CLI Tool
- Command-line interface để quản lý và tương tác với backend

## 📋 Yêu cầu hệ thống

### Cần cài đặt:
- **Go** 1.25.1 hoặc cao hơn
- **Node.js** 18+ và npm/yarn/pnpm
- **Git**

## 🚀 Hướng dẫn chạy project

### 1️⃣ Clone repository

```bash
git clone <repository-url>
cd mangahub_
```

### 2️⃣ Chạy Backend

#### Bước 1: Di chuyển vào thư mục backend
```bash
cd backend
```

#### Bước 2: Download dependencies
```bash
go mod download
```

#### Bước 3: Chạy migration để tạo database
```bash
go run cmd/migrate/main.go
```

#### Bước 4: Chạy các server

**Chạy HTTP API Server (Port 8080):**
```bash
go run cmd/api-server/main.go
```

**Hoặc chạy gRPC Server (Port 50051):**
```bash
go run cmd/grpc-server/main.go
```

**Hoặc chạy TCP Server (Port 9000):**
```bash
go run cmd/tcp-server/main.go
```

**Hoặc chạy UDP Server (Port 9001):**
```bash
go run cmd/udp-server/main.go
```

**Hoặc chạy WebSocket Server (Port 8081):**
```bash
go run cmd/ws-server/main.go
```

> **Lưu ý:** Bạn có thể chạy nhiều server cùng lúc bằng cách mở nhiều terminal.

### 3️⃣ Chạy Frontend

#### Bước 1: Mở terminal mới và di chuyển vào thư mục frontend
```bash
cd frontend
```

#### Bước 2: Cài đặt dependencies
```bash
npm install
# hoặc
yarn install
# hoặc
pnpm install
```

#### Bước 3: Chạy development server
```bash
npm run dev
# hoặc
yarn dev
# hoặc
pnpm dev
```

Frontend sẽ chạy tại: **http://localhost:3000**

### 4️⃣ Chạy CLI Tool (Optional)

#### Bước 1: Di chuyển vào thư mục cli
```bash
cd cli
```

#### Bước 2: Download dependencies
```bash
go mod download
```

#### Bước 3: Build CLI
```bash
go build -o mangahub-cli main.go
```

#### Bước 4: Chạy CLI commands
```bash
# Windows
.\mangahub-cli --help

# Linux/Mac
./mangahub-cli --help
```

## 🔧 Cấu hình

### Backend Configuration
Backend có thể được cấu hình thông qua file `.env` hoặc environment variables. Xem thêm tại `backend/internal/config/`.

### Frontend Configuration
Frontend configuration có thể được tìm thấy tại `frontend/config/env.ts`.

## 📁 Cấu trúc thư mục

```
mangahub_/
├── backend/           # Go backend với nhiều server
│   ├── cmd/          # Entry points cho các server
│   ├── internal/     # Application logic
│   ├── domain/       # Domain models
│   ├── db/           # Database và migrations
│   └── proto/        # Protocol buffers
├── frontend/         # Next.js frontend
│   ├── app/          # Next.js App Router
│   ├── components/   # React components
│   └── service/      # API services
├── cli/              # Command-line tool
└── docs/             # Documentation
```

## 🌐 Endpoints mặc định

- **Frontend**: http://localhost:3000
- **HTTP API**: http://localhost:8080
- **gRPC**: localhost:50051
- **TCP**: localhost:9000
- **UDP**: localhost:9001
- **WebSocket**: ws://localhost:8081

## 📖 Tài liệu chi tiết

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Kiến trúc hệ thống chi tiết
- [cli-config.md](docs/cli-config.md) - Hướng dẫn CLI
- [Use Cases](docs/use-cases/) - Các use case cụ thể

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm run test
```

## 📝 Các lệnh hữu ích

### Backend
```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Build tất cả servers
go build ./cmd/...
```

### Frontend
```bash
# Build production
npm run build

# Start production server
npm run start

# Lint code
npm run lint
```

## 🤝 Contributing

Mọi đóng góp đều được hoan nghênh! Vui lòng tạo issue hoặc pull request.

## 📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 🆘 Troubleshooting

### Backend không kết nối được database
- Đảm bảo đã chạy migration: `go run cmd/migrate/main.go`
- Kiểm tra file database được tạo trong thư mục backend

### Frontend không gọi được API
- Đảm bảo backend HTTP server đang chạy trên port 8080
- Kiểm tra cấu hình API endpoint trong `frontend/config/env.ts`

### Port đã được sử dụng
- Thay đổi port trong code hoặc dừng process đang dùng port đó

## 📞 Liên hệ

Nếu có câu hỏi hoặc vấn đề, vui lòng tạo issue trên GitHub repository.

---

**Happy Coding! 🎉**
