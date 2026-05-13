# GoProject — User, Team & Asset Management

A full-stack web application for managing users, teams, folders, and notes with role-based access control and asset sharing.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, Gin, MySQL |
| Frontend | React 19, Vite, TanStack Query, Zustand, Tailwind CSS |
| Auth | JWT (HS256, 24h expiry) |
| Password | bcrypt |

## Architecture

```
Frontend (React + Vite) — http://localhost:5173
        │
        │  REST API / JSON
        ▼
Backend (Go + Gin)       — http://localhost:8080/api
        │
        ▼
   MySQL Database
```

## Features Implemented

### Stage 1 — Identity & Auth ✅
- User registration with role (`manager` / `member`)
- Login / Logout with JWT
- In-memory token revocation on logout
- Role-based middleware (`AuthRequired`, `RequireManager`)

### Stage 2 — Team Management ✅
- Manager creates teams and assigns roles (OWNER / MANAGER / MEMBER)
- Add / remove team members
- List all teams the current user belongs to
- Role hierarchy enforced: only OWNER can remove managers or delete team

### Stage 3 — Asset Management ✅
- Folders owned by users (full CRUD)
- Notes inside folders (full CRUD)
- Access control: owner → full; shared write → edit; shared read → read-only
- Manager oversight: managers have implicit read-only on all assets of team members

### Stage 4 — Sharing ✅
- Share a folder or note with any user by email (Read or Write)
- Folder inheritance: sharing a folder automatically shares all its notes
- Revoke access at any time
- Manager oversight: read-only by default unless explicitly shared with write

### Stage 5 — Bulk User Import ✅
- `POST /api/users/import` — accepts a `.csv` file
- Concurrent processing via goroutine worker pool (5 workers, channels, WaitGroup)
- Returns summary: `{ succeeded, failed, errors[] }`
- CSV format: `username,email,password[,role]` — role defaults to `member`

## Project Structure

```
GoProject/
├── Backend/          # Go REST API
│   ├── api/          # Route registration
│   ├── cmd/          # Entry point (main.go)
│   ├── internal/
│   │   ├── config/       # DB + CORS config
│   │   ├── handler/      # HTTP handlers (7 files)
│   │   ├── middleware/   # Auth, rate limiting, CORS
│   │   ├── model/        # Domain models
│   │   ├── repository/   # Raw SQL (no ORM)
│   │   └── service/      # Business logic
│   └── package/
│       ├── dtorequest/   # Input DTOs
│       └── utils/        # Validators
│
└── Frontend/GoProject/   # React SPA
    └── src/
        ├── pages/        # DashboardPage, TeamPage, ImportPage, ...
        ├── shared/
        │   ├── components/   # Button, Card, Table, ...
        │   ├── hooks/        # useTeams, useFolders, useNotes, ...
        │   ├── layouts/      # DashboardLayout, AuthLayout
        │   └── services/     # API clients (axios-based)
        └── stores/           # Zustand (auth, ui)
```

## Quick Start

### 1. Database

Start MySQL via Docker:

```bash
cd Backend
docker-compose -f db.yaml up -d
```

Run schema migrations (see `Backend/migration_*.sql`).

### 2. Backend

```bash
cd Backend
cp .env.example .env   # fill in DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_NAME, JWT_SECRET
go mod tidy
go run ./cmd
# Server starts on :8080
```

### 3. Frontend

```bash
cd Frontend/GoProject
npm install
npm run dev
# App starts on :5173
```

## Known Limitations

- Token revocation list is in-memory — cleared on server restart
- No refresh token support
- No database migration framework (tables must be created manually)
- No automated test suite
