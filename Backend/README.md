# Backend

Go REST API (Gin) backed by MySQL, Redis cache, RabbitMQ events, JWT auth, and structured logging with zerolog. Passwords hashed with bcrypt; data access via raw `database/sql` (no ORM).

## Architecture

Strict three-layer pattern — **Handler → Service → Repository → MySQL** — augmented with a read-through Redis cache and a RabbitMQ pub/sub fabric for audit logging, notifications, and cache invalidation.

```
cmd/main.go              # Entry point: wires DB, cache, broker, services; listens on :8080
api/
  MainRouting.go         # Route registration + middleware
  swagger.yaml           # OpenAPI 3.0 spec (full API reference)
  integration_test.go    # Router-level integration tests
internal/
  config/                # DB pool, CORS, cache, rate limit, RabbitMQ, log config
  handler/               # HTTP layer (parse → service → respond)
    AuthHandler.go
    UserHandler.go
    TeamHandler.go
    FolderHandler.go
    NoteHandler.go
    SharingHandler.go
    ImportHandler.go
  middleware/            # JWT auth, role check, rate limit (IP + per-user), CORS, request ID, logging
  model/                 # User, Team, TeamMember, Folder, Note, Permission, ImportTask, AuditLog
  repository/            # Raw SQL with context.Context
  service/               # Business logic + *_test.go (sqlmock)
    AuthService.go
    UserService.go       # includes async CSV import
    TeamManagementService.go
    FolderService.go
    NoteService.go
    SharingService.go
  cache/                 # Redis wrappers: ACL, asset metadata, team members; noop fallback
  rabbitmq/              # AMQP connection / channel / publish / consume
  consumer/              # Audit log, notification, cache-invalidator consumers
package/
  dtorequest/            # Input DTOs with validator tags
  event/                 # Event publisher (RabbitMQ) + noop fallback
  utils/                 # Custom validator registration
schema.sql               # Full initial schema
migration_*.sql          # Migration scripts (already folded into schema.sql)
Dockerfile               # Multi-stage build, alpine runtime (~30 MB)
db.yaml                  # Docker compose for MySQL only (dev convenience)
queue.yaml               # Docker compose for RabbitMQ only (dev convenience)
```

## Environment Setup

Create `Backend/.env`:

```env
# MySQL
DB_HOST=localhost
DB_PORT=3311
DB_USERNAME=user
DB_PASSWORD=1234
DB_NAME=miniproject_database

# JWT (required — server panics if missing)
JWT_SECRET=your_secret_here

# RabbitMQ (optional — missing/down degrades to a noop publisher; events are dropped)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Redis (optional — missing/timeout degrades to a noop cache; reads fall through to DB)
# Defaults to localhost:6379 when unset
REDIS_ADDR=localhost:6379
```

Only `JWT_SECRET` is required. Redis and RabbitMQ outages do **not** block startup — the app logs a warning and falls back to noop implementations.

## Run

```bash
# From Backend/
go mod tidy
go run ./cmd            # → :8080

# Run tests (sqlmock — no live DB required)
go test ./...

# Bring up MySQL / RabbitMQ separately for local dev
docker compose -f db.yaml up -d
docker compose -f queue.yaml up -d
```

For the full stack (API + DB + Redis + RabbitMQ + Grafana/Loki/Promtail) use the root `docker-compose.yml`.

## Database Schema

See `schema.sql` for the complete schema (migrations already folded in). Summary of the 9 tables:

| Table | Purpose |
| --- | --- |
| `users` | Accounts (UUID PK, unique username/email, bcrypt hash, role) |
| `teams` | Teams (unique name, `owner_id` UUID) |
| `team_members` | Membership with in-team role (`owner`, `manager`, `member`) |
| `folders` | Folders (`folder_id` PK, `owner_id`) |
| `notes` | Notes (`note_id` PK, `folder_id`, `owner_id`) |
| `permissions` | ACL (`asset_type` ∈ `folder`/`note`, `permission_type` ∈ `read`/`write`) |
| `import_tasks` | Async CSV import state (status, progress, error_log JSON) |
| `audit_logs` | Event sink from RabbitMQ (topic, event_type, payload JSON) |

> Note: folder PK is `folder_id`, note PK is `note_id`. Timestamp columns use `DATETIME` (not `TIMESTAMP`).

No migration framework yet — run `schema.sql` manually against a fresh database.

## API Reference

Base URL: `http://localhost:8080/api`. Every 🔒 endpoint requires `Authorization: Bearer <token>`. See `api/swagger.yaml` for full request/response schemas.

### Health

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Returns 200 when the server is up |

### Auth

| Method | Path | Description |
| --- | --- | --- |
| POST | `/auth/login` | Returns `{ token, user }`; JWT HS256, 24 h TTL |
| POST | `/auth/logout` 🔒 | Adds the token's JTI to an in-memory blocklist |

### Users

| Method | Path | Description |
| --- | --- | --- |
| POST | `/users/register` | Create a user; `role` ∈ `manager`/`member` |
| GET | `/users` 🔒 | List all users |
| POST | `/users/import` 🔒 | Upload CSV `username,email,password[,role]`; returns `202` with `task_id` |
| GET | `/import-tasks/:id` 🔒 | Poll status of an import task |

`/users/import` runs asynchronously through a 5-worker goroutine pool, flushing progress every 500 rows or 2 seconds. Upload capped at 10 MB.

### Teams

| Method | Path | Description |
| --- | --- | --- |
| GET | `/teams` 🔒 | Teams the caller belongs to |
| POST | `/teams` 🔒 (manager) | Create a team — the creator becomes `OWNER` automatically |
| POST | `/teams/:teamName/members` 🔒 (manager) | Add a member by username |
| DELETE | `/teams/:teamName/members/:memberName` 🔒 (manager) | Remove a member |
| DELETE | `/teams/:teamName` 🔒 (manager) | Delete the team — `OWNER` only |

### Folders & Notes

| Method | Path | Description |
| --- | --- | --- |
| GET / POST | `/folders` 🔒 | List / create folders |
| GET / PUT / DELETE | `/folders/:id` 🔒 | Read / update / delete a folder (ACL-checked) |
| GET / POST | `/folders/:id/notes` 🔒 | List / create notes inside a folder |
| GET / PUT / DELETE | `/notes/:id` 🔒 | Read / update / delete a note |

### Sharing

| Method | Path | Description |
| --- | --- | --- |
| POST | `/share` 🔒 | Grant `read`/`write` access to a user (by email) on a folder or note |
| DELETE | `/share` 🔒 | Revoke access |

Sharing or revoking a folder cascades to every note currently inside it, then publishes a RabbitMQ event so Redis ACL entries are invalidated across all instances.

## Access Control (4-stage check)

| # | Stage | Outcome |
| --- | --- | --- |
| 1 | Owner | Full access |
| 2 | Explicit ACL (Redis 30 min TTL → fallback to DB) | `read` or `write` per `permission_type` |
| 3 | Team Manager of the owner | Read-only (implicit — no explicit share required) |
| 4 | Default | HTTP 403 |

Owner always passes; every other stage hits the cache before falling back to MySQL.

## Rate Limiting

In-memory, two tiers, zero external dependencies:

| Endpoint | Limit |
| --- | --- |
| `POST /auth/login` | 10,000 req/min/IP |
| `POST /users/register` | 10,000 req/min/IP |
| `POST /users/import` | 5 req/min/IP **and** 3 req/10 min per user |

## Caching & Events

- **Redis** (read-through, write-on-miss):
  - ACL (folder/note ↔ user → permission), 30 min TTL
  - Asset metadata (folder/note), 1 h TTL
  - Team membership (team_id → user_ids, user_id → team_ids), 30 min TTL
  - Any outage or timeout falls back to `NoopCache` — requests are never blocked on the cache.
- **RabbitMQ** publisher feeds:
  - `audit.events` → `AuditConsumer` writes to the `audit_logs` table
  - `notification.*` → `NotificationConsumer` (structured log)
  - `cache.invalidate.*` → `CacheInvalidator` evicts the corresponding Redis entries
- If the broker is down the publisher becomes `NoopPublisher` and consumers are not started until the next successful boot.

## Observability

The backend emits JSON logs via **zerolog** to stdout. Each request gets a `request_id` (assigned by middleware) and carries level, latency, and route. In the full-stack compose, Promtail → Loki → Grafana ship and visualize the logs.

## Key Patterns

- **Adding an endpoint**: DTO (`package/dtorequest`) → repository method → service method → handler method → route in `api/MainRouting.go`.
- **Validation**: `c.ShouldBindJSON(&req)` plus `binding:"required,..."` struct tags; custom validators registered via `utils.RegisterValidators()`.
- **Context**: every repository method takes `context.Context` first; handlers pass `c.Request.Context()`.
- **Errors**: MySQL error 1062 → duplicate-key domain errors (`ErrEmailAlreadyExists`, etc.); `sql.ErrNoRows` → `ErrNotFound` variants.
- **Service wiring**: builder pattern — `service.NewX(...).WithPublisher(pub).WithCache(c1, c2)`. Caches and publishers are optional; nil-safe wrappers cover the noop case.

## Tests

Service-layer unit tests use `go-sqlmock` — no live DB required:

```bash
go test ./...
```

Existing test files: `auth_service_test.go`, `user_service_test.go`, `team_management_service_test.go`, `folder_service_test.go`, `note_service_test.go`, `sharing_service_test.go`, `internal/middleware/auth_test.go`, `api/integration_test.go`.

## Known Limitations

| Area | Current state | Path forward |
| --- | --- | --- |
| Token revocation | In-memory JTI map; resets on restart | Redis-backed blocklist with TTL |
| Rate limiter | Single-process in-memory | Redis-based for multi-instance |
| Folder share inheritance | Point-in-time; notes added after sharing are not covered | On-write / on-read resolution |
| Pagination | All list endpoints return full result sets | `limit`/`offset` or cursor-based |
| Schema migrations | Manual `.sql` files | `golang-migrate` or Atlas |
| Refresh tokens | Not implemented | Refresh + rotation |
