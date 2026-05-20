# GoProject — User, Team & Asset Management

A full-stack web application for managing users, teams, folders, and notes with role-based access control, asset sharing, and async bulk import.

---

## Tech Stack

### Backend

| Component | Technology | Version |
|---|---|---|
| Language | Go | 1.26.2 |
| HTTP Framework | Gin | 1.12.0 |
| Database | MySQL | via Docker |
| Driver | go-sql-driver/mysql | 1.10.0 |
| Auth | golang-jwt/jwt | v5.3.1 |
| Password | golang.org/x/crypto (bcrypt) | 0.49.0 |
| Validation | go-playground/validator | v10.30.2 |
| UUID | google/uuid | 1.6.0 |
| Config | joho/godotenv | 1.5.1 |
| Test mock | DATA-DOG/go-sqlmock | 1.5.2 |

### Frontend

| Component | Technology | Version |
|---|---|---|
| UI Library | React | 19 |
| Build Tool | Vite | 8 |
| Server State | TanStack Query | v5.100 |
| Client State | Zustand | v5 |
| Styling | Tailwind CSS | v4 |
| Router | React Router | v7 |
| Forms | React Hook Form + Zod | v7 + v4 |
| HTTP Client | Axios | 1.15 |

---

## Architecture

```
Browser (React SPA)          — http://localhost:5173
        │
        │  REST / JSON over HTTP
        ▼
Go + Gin API Server          — http://localhost:8080/api
        │
        │  database/sql (raw SQL, no ORM)
        ▼
MySQL Database               — localhost:3311 (Docker)
```

### Backend Layer Model

```
Handler  →  Service  →  Repository  →  MySQL
```

- **Handler** — parses HTTP request, calls service, writes JSON response. No business logic.
- **Service** — all business rules, authorization checks, cross-entity coordination.
- **Repository** — raw SQL queries only; returns domain models. No logic.

### Frontend Module Map

```
src/
├── pages/               # Route-level components (one folder per page)
│   ├── LoginPage/       # Login form + useLogin hook + Zod schema
│   ├── RegisterPage/    # Register form + useRegister hook + Zod schema
│   ├── DashboardPage/   # User list table + summary cards
│   ├── TeamPage/        # Team workspace + management panel
│   ├── ImportPage/      # Async multi-file CSV import + live polling
│   └── ProfilePage/     # Current user profile
├── shared/
│   ├── components/      # Button, Card, Table, Input, Select, Toast,
│   │                    # RoleBadge, EmptyState, LoadingSkeleton, DashboardStats
│   ├── hooks/           # useUsers, useTeams, useLogout
│   ├── layouts/         # DashboardLayout, AuthLayout
│   ├── services/        # Axios instance + API modules per domain
│   │                    # (authApi, usersApi, teamsApi, foldersApi, notesApi, sharingApi)
│   ├── constants/       # roles.js, routes.js
│   ├── types/           # user.js (normalizer)
│   └── utils/           # ProtectedRoute, PublicRoute, storage, formatDate
└── stores/              # Zustand stores: authStore, uiStore, importStore
```

---

## Features

### Auth
- Registration with role (`manager` / `member`); JWT issued immediately on success
- Login / Logout; logout invalidates the token via in-memory JTI revocation list
- `AuthRequired` middleware validates JWT on every protected route
- `RequireManager` middleware gates manager-only endpoints

### Team Management
- Manager creates a team; creator is automatically assigned `OWNER`
- Three-tier team hierarchy: `OWNER` > `MANAGER` > `MEMBER`
  - Only `OWNER` can remove managers or delete the team
  - `OWNER` and `MANAGER` can add / remove members
- List all teams the authenticated user belongs to

### Asset Management (Folders & Notes)
- Full CRUD on folders and notes; access is always checked against the requester
- Permission cascade: `OWNER` → `WRITE` share → `READ` share → team manager (read-only oversight)
- Folder sharing propagates permissions to all notes currently in the folder
- Revoking folder access cascades to contained notes

### Async Bulk Import
- `POST /api/users/import` accepts a `.csv` file and returns a `task_id` immediately (HTTP 202)
- A goroutine worker pool (5 workers + buffered channels) processes rows concurrently
- Progress is persisted to DB on every 500 rows **or** every 2 seconds (dual-trigger)
- `GET /api/import-tasks/:id` returns live status, progress, and error details
- Frontend polls every 2 s using TanStack Query `refetchInterval`; stops automatically when `completed` or `failed`
- Multiple files can be queued simultaneously; task list survives tab navigation via Zustand + `sessionStorage`

### Middleware Stack
| Middleware | Detail |
|---|---|
| CORS | Configurable via `CorsConfig`; applied globally |
| Request ID | Attaches a UUID to every request for tracing |
| Logging | Logs method, path, status, latency per request |
| Auth | JWT validation + JTI revocation check |
| Rate Limiting | Two-tier: global IP limiter + per-endpoint limiter with independent windows and cleanup goroutine |

---

## API Reference

| Method | Path | Auth | Role | Description |
|---|---|---|---|---|
| POST | `/api/auth/login` | — | — | Login |
| POST | `/api/auth/logout` | ✓ | — | Logout (revoke token) |
| POST | `/api/users/register` | — | — | Register |
| GET | `/api/users` | ✓ | manager | List all users |
| POST | `/api/users/import` | ✓ | manager | Start async CSV import |
| GET | `/api/import-tasks/:id` | ✓ | — | Poll import task status |
| GET | `/api/teams` | ✓ | — | List teams for current user |
| POST | `/api/teams` | ✓ | manager | Create team |
| POST | `/api/teams/:name/members` | ✓ | manager | Add member |
| DELETE | `/api/teams/:name/members/:member` | ✓ | manager | Remove member |
| DELETE | `/api/teams/:name` | ✓ | manager | Delete team |
| POST | `/api/folders` | ✓ | — | Create folder |
| GET | `/api/folders` | ✓ | — | List own folders |
| GET | `/api/folders/:id` | ✓ | — | Get folder |
| PUT | `/api/folders/:id` | ✓ | — | Update folder |
| DELETE | `/api/folders/:id` | ✓ | — | Delete folder |
| POST | `/api/folders/:id/notes` | ✓ | — | Create note in folder |
| GET | `/api/folders/:id/notes` | ✓ | — | List notes in folder |
| GET | `/api/notes/:id` | ✓ | — | Get note |
| PUT | `/api/notes/:id` | ✓ | — | Update note |
| DELETE | `/api/notes/:id` | ✓ | — | Delete note |
| POST | `/api/share` | ✓ | — | Share asset by email |
| DELETE | `/api/share` | ✓ | — | Revoke access |

---

## Project Structure

```
GoProject/
├── test.csv                          # Sample import file
│
├── Backend/
│   ├── cmd/main.go                   # Entry point: wires all layers, starts on :8080
│   ├── api/MainRouting.go            # Route registration and middleware attachment
│   ├── db.yaml                       # Docker Compose for MySQL
│   ├── migration_add_role.sql        # Schema migration: role column
│   ├── migration_add_owner_to_teams.sql
│   ├── go.mod / go.sum
│   └── internal/
│       ├── config/
│       │   ├── DbConfig.go           # DSN builder from env vars
│       │   └── CorsConfig.go         # CORS policy
│       ├── handler/
│       │   ├── AuthHandler.go
│       │   ├── UserHandler.go
│       │   ├── ImportHandler.go      # Async import: saves temp file, spawns goroutine
│       │   ├── TeamHandler.go
│       │   ├── FolderHandler.go
│       │   ├── NoteHandler.go
│       │   └── SharingHandler.go
│       ├── middleware/
│       │   ├── auth.go               # JWT validation + in-memory revocation
│       │   ├── cors.go
│       │   ├── ratelimit.go          # Custom IP + per-endpoint rate limiter
│       │   ├── logging.go
│       │   └── requestid.go
│       ├── model/
│       │   ├── User.go
│       │   ├── Team.go / TeamMember.go
│       │   ├── Folder.go / Note.go
│       │   ├── Permission.go
│       │   └── ImportTask.go
│       ├── repository/
│       │   ├── Database.go           # Connection pool (25 max open, 5 idle)
│       │   ├── UserRepository.go
│       │   ├── TeamRepository.go / TeamMemberRepository.go
│       │   ├── FolderRepository.go / NoteRepository.go
│       │   ├── PermissionRepository.go
│       │   └── ImportTaskRepository.go
│       └── service/
│           ├── AuthService.go
│           ├── UserService.go        # Register + async CSV import pipeline
│           ├── TeamManagementService.go
│           ├── FolderService.go      # Permission cascade logic
│           ├── NoteService.go        # Four-stage access check
│           └── SharingService.go     # Folder → note inheritance
│
└── Frontend/GoProject/
    ├── index.html
    ├── vite.config.js
    ├── package.json
    └── src/
        ├── App.jsx                   # Route definitions
        ├── main.jsx                  # React root + QueryClient + Router
        ├── pages/
        │   ├── LoginPage/            # LoginForm, useLogin, authSchemas
        │   ├── RegisterPage/         # RegisterForm, useRegister, authSchemas
        │   ├── DashboardPage/        # UserTable, UserSummaryCards
        │   ├── TeamPage/             # TeamWorkspace, TeamManagementPanel
        │   ├── ImportPage/           # Multi-file async import + live polling
        │   ├── ProfilePage/
        │   └── NotFoundPage/
        ├── shared/
        │   ├── components/           # Button, Card, Table, Input, Select,
        │   │                         # Toast, RoleBadge, EmptyState,
        │   │                         # LoadingSkeleton, DashboardStats
        │   ├── hooks/                # useUsers, useTeams, useLogout
        │   ├── layouts/              # DashboardLayout, AuthLayout
        │   ├── services/             # authApi, usersApi, teamsApi,
        │   │                         # foldersApi, notesApi, sharingApi,
        │   │                         # axios instance, interceptor, apiError
        │   ├── constants/            # roles, routes
        │   ├── types/                # user normalizer
        │   └── utils/                # ProtectedRoute, PublicRoute, storage, formatDate
        └── stores/
            ├── authStore.js          # JWT token + user; persisted to localStorage
            ├── uiStore.js            # Sidebar, theme
            └── importStore.js        # Task ID list; persisted to sessionStorage
```

---

## Tech Stack — Trade-offs

### Go + Gin
**Advantages**
- Compiled binary; low memory footprint and fast cold start compared to Node/Python equivalents
- Goroutine-based concurrency maps naturally to the worker-pool import pipeline
- Strong typing catches entire classes of bugs at compile time

**Trade-offs**
- More boilerplate than Express or FastAPI for simple CRUD endpoints
- No built-in ORM means query strings are scattered across repository files; schema refactors require manual updates in multiple places

### Raw SQL (`database/sql`, no ORM)
**Advantages**
- Full control over query shape and index usage; no N+1 surprises hidden behind an abstraction
- Zero runtime reflection overhead

**Trade-offs**
- No automatic schema migrations; `migration_*.sql` files must be applied manually
- Typos in column names are only caught at runtime, not compile time
- Joining or changing a table name requires grep-and-replace across repository files

### JWT (HS256, 24h, in-memory revocation)
**Advantages**
- Stateless verification: any instance can validate a token without a DB or cache lookup
- Simple to implement and reason about for single-instance deployments

**Trade-offs**
- The JTI revocation map is in-memory; server restart re-validates all previously revoked tokens until they expire naturally
- 24h expiry is long for sensitive operations; without refresh tokens, users must log in again after expiry
- A single shared `JWT_SECRET` means all tokens are invalidated if the secret rotates

### Custom In-Memory Rate Limiter
**Advantages**
- Zero external dependencies; works out of the box
- Per-endpoint and global limits configurable independently at startup

**Trade-offs**
- State is local to one process; does not share limits across multiple backend instances
- All counters reset on restart

### TanStack Query v5 (frontend)
**Advantages**
- Declarative server-state management; handles caching, background refetch, and `refetchInterval` (used for import polling) with minimal code
- Deduplicates in-flight requests automatically

**Trade-offs**
- Adds ~13 KB (gzip) to the bundle; overkill if the app only has a handful of API calls
- v5 breaking changes (function signatures for `refetchInterval`) require attention when upgrading

### Zustand v5 (frontend)
**Advantages**
- Minimal boilerplate compared to Redux; store slices are plain functions
- `persist` middleware integrates cleanly with `localStorage` / `sessionStorage`

**Trade-offs**
- No built-in devtools comparable to Redux DevTools (requires a separate plugin)
- No enforced immutability; accidental mutation bypasses subscribers silently

### React Hook Form + Zod
**Advantages**
- Uncontrolled inputs mean zero re-renders per keystroke; good for large forms
- Zod schemas are reusable as TypeScript/runtime validators and double as documentation

**Trade-offs**
- Additional learning surface for developers who already know simpler validation libraries
- Error messages are in English by default; i18n requires extra wiring

---

## Quick Start

### Prerequisites
- Go ≥ 1.21, Docker, Node.js ≥ 18

### 1. Database

```bash
cd Backend
docker-compose -f db.yaml up -d
```

Apply the schema migrations manually:

```bash
mysql -u user -p miniproject_database < migration_add_role.sql
mysql -u user -p miniproject_database < migration_add_owner_to_teams.sql
```

### 2. Backend

```bash
cd Backend
cp .env.example .env   # or create .env manually
go mod tidy
go run ./cmd
# Listening on :8080
```

Required environment variables:

```env
DB_HOST=localhost
DB_PORT=3311
DB_USERNAME=user
DB_PASSWORD=1234
DB_NAME=miniproject_database
JWT_SECRET=<any strong random string>   # required — server panics on startup if missing
```

### 3. Frontend

```bash
cd Frontend/GoProject
npm install
npm run dev
# Listening on :5173
```

### Running Tests

```bash
cd Backend
go test ./...
```

Tests use `go-sqlmock` for repository-level mocking and cover service logic for all five service packages.

---

## Known Limitations & Areas for Improvement

| Area | Current State | Suggested Improvement |
|---|---|---|
| Schema migrations | Manual `.sql` files | Adopt `golang-migrate` or Atlas for versioned, repeatable migrations |
| Token revocation | In-memory JTI map; resets on restart | Move to Redis with TTL equal to token expiry |
| Refresh tokens | Not implemented | Add short-lived access tokens (15 min) + long-lived refresh tokens |
| Rate limiting | Single-process in-memory | Replace with Redis-based limiter (e.g. `go-redis/redis_rate`) for multi-instance support |
| Import progress | HTTP polling every 2 s | Replace with Server-Sent Events or WebSocket for true push-based progress |
| Folder share inheritance | Point-in-time snapshot | Notes added to a folder after sharing are not covered automatically |
| Pagination | All list endpoints return full results | Add `limit` / `offset` or cursor-based pagination |
| Structured logging | Mix of `fmt.Print` and `log.Printf` | Adopt `slog` (stdlib) or `zap` for JSON-structured, levelled logging |
| API versioning | No versioning; all routes under `/api` | Prefix routes with `/api/v1` to allow non-breaking evolution |
| Error messages | English only | Add i18n layer for client-facing error strings |
| `revokedBy` audit field | Accepted but not persisted in `SharingService` | Store in a `permissions_audit` table for compliance |
| Connection pool | Hardcoded (25 max open, 5 idle) | Make pool size configurable via env vars |
