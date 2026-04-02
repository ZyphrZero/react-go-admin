# React Go Admin

一个基于 Go + Gin + GORM + React + Vite + Bun 的后台管理系统。

## Overview
- Backend: Go 1.26+, Gin, GORM, slog
- Frontend: React 19, Vite 8, shadcn/ui, Bun
- Database: SQLite (default), MySQL, PostgreSQL
- Runtime model: backend starts from `app/main.go`, and startup can auto-run migration + seed

## Preview
![login_light_dark_diagonal_preview.png](https://cdn.pixelpunk.cc/f/01c51f0211f647ef/login_light_dark_diagonal_preview.png)
![dashboard_light_dark_diagonal_preview.png](https://cdn.pixelpunk.cc/f/4735f2fde4244262/dashboard_light_dark_diagonal_preview.png)

## Project Layout

```text
.
├── app/
│   ├── main.go
│   └── internal/
│       ├── config/
│       ├── core/
│       ├── framework/
│       ├── http/router/
│       ├── modules/
│       ├── migrate/
│       ├── platform/
│       └── seed/
├── web/
├── deploy/
├── go.mod
└── .env.example
```

## Local Development

### 1) Prerequisites
- Go 1.26+
- Bun 1.3+

### 2) Environment

```bash
cp .env.example .env
```

PowerShell:

```powershell
Copy-Item .env.example .env
```

### 3) Start Backend

```bash
go mod download
go run ./app
```

Default startup behavior:
- auto-apply schema migration
- auto-seed baseline data (roles/admin/default permissions/API catalog)
- keep `/api/v1/init/checkdb` and `/api/v1/init/initdb` available when `ENABLE_INIT_API=true`

Disable startup bootstrap:

```bash
DISABLE_AUTO_MIGRATE=true go run ./app
```

PowerShell:

```powershell
$env:DISABLE_AUTO_MIGRATE = "true"
go run ./app
```

Disable init endpoints:

```powershell
$env:ENABLE_INIT_API = "false"
go run ./app
```

### 4) Start Frontend

```bash
cd web
bun install
bun run dev
```

## Endpoints
- API: `http://127.0.0.1:9999`
- Health: `http://127.0.0.1:9999/health`
- Frontend dev: `http://127.0.0.1:5173`

## Testing

Backend:

```bash
go test ./...
```

Frontend:

```bash
cd web
bun run lint
bun run build
```

## Docker Deployment

Deployment files:
- `deploy/.env.example`
- `deploy/docker-compose.yml`
- `deploy/install.sh`

Quick start:

```bash
cd deploy
cp .env.example .env
chmod +x install.sh
./install.sh install
```

Useful commands:

```bash
./install.sh upgrade
./install.sh status
./install.sh logs
./install.sh down
```

## License
[MIT](./LICENSE)
