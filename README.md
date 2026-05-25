# GoProject — Collaborative Note & Team Management Platform

A full-stack platform built on **Go + React** that lets teams manage users, folders, notes, and shared assets — with real-time async import, fine-grained access control, and built-in observability. Everything ships in a single `docker compose up`.

---

## System Architecture

```mermaid
graph TB
    subgraph Client["🖥️  Client Layer"]
        UI["React 19 SPA\n(Vite · TanStack Query · Zustand · Tailwind)"]
    end

    subgraph Gateway["🌐  API Gateway  :8080"]
        MW["Middleware\n(Auth · Rate Limit · CORS · Logger)"]
        Router["Gin Router\n22 REST endpoints"]
    end

    subgraph Core["⚙️  Application Core"]
        direction TB
        H["Handlers\n(Auth · User · Team · Folder · Note · Sharing · Import)"]
        S["Services\n(Business Logic)"]
        R["Repositories\n(Raw database/sql — no ORM)"]
    end

    subgraph Infra["🗄️  Infrastructure"]
        direction LR
        DB[("MySQL 8.4\n9 tables")]
        Cache[("Redis 7\nACL · Team · Metadata\nTTL 30 min / 1 h")]
        MQ["RabbitMQ 3\nCache Invalidation\nEvents"]
    end

    subgraph Observability["📊  Observability Stack"]
        direction LR
        PT["Promtail 3.5.6\nLog Shipper"]
        LK["Loki 3.5.6\nLog Aggregation"]
        GF["Grafana 11.6.1\nDashboard  :3000"]
    end

    UI -->|REST / JSON| MW
    MW --> Router
    Router --> H
    H --> S
    S --> R
    R -->|SQL| DB
    S <-->|Cache read/write| Cache
    S -->|Publish events| MQ
    MQ -->|Subscribe & invalidate| Cache

    Core -->|structured JSON logs| PT
    PT -->|push| LK
    LK --> GF

    style Client fill:#dbeafe,stroke:#3b82f6,color:#1e3a8a
    style Gateway fill:#ede9fe,stroke:#7c3aed,color:#3b0764
    style Core fill:#dcfce7,stroke:#16a34a,color:#14532d
    style Infra fill:#fef9c3,stroke:#ca8a04,color:#713f12
    style Observability fill:#ffe4e6,stroke:#e11d48,color:#881337
```

---

## Request Lifecycle

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as Middleware
    participant H as Handler
    participant S as Service
    participant Redis as Redis Cache
    participant DB as MySQL

    C->>MW: HTTP Request + JWT
    MW->>MW: Validate JWT & check JTI blocklist
    MW->>MW: Rate limit check (IP / user)
    MW->>H: ctx with userID

    H->>S: Call service method
    S->>Redis: Cache lookup
    alt Cache hit
        Redis-->>S: Cached data ✓
    else Cache miss
        S->>DB: SQL query
        DB-->>S: Result rows
        S->>Redis: Write-through cache
    end
    S-->>H: Domain struct
    H-->>C: JSON response
```

---

## Access Control Flow

```mermaid
flowchart TD
    REQ(["🔒 Resource Request"]) --> A{"Is requester\nthe Owner?"}

    A -->|Yes| GRANT(["✅ Full Access"])

    A -->|No| B{"Explicit permission\nin Redis ACL?"}
    B -->|Hit| GRANT

    B -->|Miss| C["Query DB\n→ Populate Redis TTL 30 min"]
    C --> D{"Permission\nfound?"}
    D -->|Yes| GRANT

    D -->|No| E{"Requester is\nTeam Manager\nof owner?"}
    E -->|Yes| READONLY(["👁️ Read-Only Access"])

    E -->|No| DENY(["❌ HTTP 403 Forbidden"])

    style GRANT fill:#bbf7d0,stroke:#16a34a,color:#14532d
    style READONLY fill:#fef9c3,stroke:#ca8a04,color:#713f12
    style DENY fill:#fecaca,stroke:#dc2626,color:#7f1d1d
    style REQ fill:#dbeafe,stroke:#3b82f6,color:#1e3a8a
```

---

## Async Bulk Import Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as POST /api/users/import
    participant Pool as Worker Pool (5 goroutines)
    participant DB as MySQL
    participant Poll as GET /api/import-tasks/:id

    C->>API: Upload CSV file
    API-->>C: 202 Accepted { task_id, status: "pending" } < 5ms

    Note over Pool: Process rows concurrently
    loop Every 500 rows or 2s
        Pool->>DB: Flush batch progress
    end

    loop Poll every 2s (TanStack Query)
        C->>Poll: GET /api/import-tasks/:id
        Poll-->>C: { progress, total, status }
    end

    DB-->>Pool: All rows processed
    Pool->>DB: status = "completed" | "failed"
    C->>Poll: status = completed ✓ → stop polling
```

---

## Team Hierarchy

```mermaid
graph TD
    OWNER(["👑 OWNER\n(auto-assigned to creator)"])
    MGR(["🛡️ MANAGER"])
    MEM(["👤 MEMBER"])

    OWNER -->|"can add/remove"| MGR
    OWNER -->|"can add/remove"| MEM
    OWNER -->|"can delete team"| TEAM(["🏢 Team"])

    MGR -->|"can add/remove"| MEM
    MGR -->|"read-only oversight\nover all assets"| ASSETS(["📁 Folders & Notes"])

    MEM --> ASSETS

    style OWNER fill:#fef9c3,stroke:#ca8a04,color:#713f12
    style MGR fill:#dbeafe,stroke:#3b82f6,color:#1e3a8a
    style MEM fill:#dcfce7,stroke:#16a34a,color:#14532d
    style TEAM fill:#ede9fe,stroke:#7c3aed,color:#3b0764
    style ASSETS fill:#ffe4e6,stroke:#e11d48,color:#881337
```

---

## Folder & Note Sharing

```mermaid
flowchart LR
    subgraph Share["Share Folder"]
        SF["Share Folder\n(by email)"]
        SF --> SN["Auto-propagate\nto all notes inside"]
        SN --> PI["Publish invalidation\nevent to RabbitMQ"]
    end

    subgraph Revoke["Revoke Access"]
        RF["Revoke Folder"]
        RF --> RN["Cascade revoke\nto all notes"]
        RN --> RI["Publish invalidation\nevent to RabbitMQ"]
    end

    subgraph Cache["Cache Layer"]
        PI --> INV["Redis ACL entries\ninvalidated"]
        RI --> INV
    end

    style Share fill:#dcfce7,stroke:#16a34a,color:#14532d
    style Revoke fill:#fecaca,stroke:#dc2626,color:#7f1d1d
    style Cache fill:#fef9c3,stroke:#ca8a04,color:#713f12
```

---

## Observability Pipeline

```mermaid
graph LR
    subgraph Containers["Docker Containers"]
        BE["Backend\n(zerolog JSON)"]
        FE["Frontend\n(nginx)"]
    end

    subgraph Collection["Log Collection"]
        PT["Promtail 3.5.6\n• Docker service discovery\n• Label extraction:\n  level · request_id · timestamp"]
    end

    subgraph Storage["Log Storage"]
        LK["Loki 3.5.6\n• Index-free storage\n• Label-based queries"]
    end

    subgraph Visualization["Visualization  :3000"]
        GF["Grafana 11.6.1\n• Auto-provisioned dashboard\n• LogQL queries\n• Alerting"]
    end

    BE -->|stdout/stderr| PT
    FE -->|stdout/stderr| PT
    PT -->|HTTP push| LK
    LK --> GF

    style Containers fill:#dbeafe,stroke:#3b82f6,color:#1e3a8a
    style Collection fill:#ede9fe,stroke:#7c3aed,color:#3b0764
    style Storage fill:#dcfce7,stroke:#16a34a,color:#14532d
    style Visualization fill:#ffe4e6,stroke:#e11d48,color:#881337
```

---

## Features

### Authentication & Authorization

JWT tokens (HS256, **24-hour expiry**) are issued on login and validated by `AuthRequired` middleware on every protected route. Logout revokes the token via an in-memory JTI blocklist. Two roles — `manager` and `member` — gate endpoint access at the middleware layer.

### Team Roles & Permissions

Three-tier structure: `OWNER > MANAGER > MEMBER`. The creator is automatically assigned `OWNER`. Managers can add and remove members; only owners may remove managers or delete the team. Membership is cached in Redis (**30-minute TTL**) to avoid repeated DB joins on every access check.

### Folder & Note Management

**22 REST endpoints** across 6 domains cover the full asset lifecycle. Every read or write runs a 4-stage access check (Owner → Explicit ACL → Team Manager → Deny). Sharing a folder propagates permissions to all notes inside instantly, and revoking cascades the same way — both operations publish invalidation events to RabbitMQ so distributed cache entries are cleared.

### Async Bulk Import

`POST /api/users/import` returns HTTP 202 with a `task_id` in under 5 ms regardless of file size. A **5-worker goroutine pool** processes rows concurrently, flushing progress to the DB every **500 rows or 2 seconds**. Clients poll `GET /api/import-tasks/:id`; the React frontend uses TanStack Query `refetchInterval: 2000` and stops automatically on terminal state.

### Rate Limiting

Custom two-tier in-memory limiter — zero external dependencies:

| Endpoint | Limit |
| --- | --- |
| `POST /auth/login` | 10,000 req / min / IP |
| `POST /users/register` | 10,000 req / min / IP |
| `POST /users/import` | 5 req / min / IP · 3 req / 10 min per user |

### Observability

All container logs are shipped by **Promtail 3.5.6** to **Loki 3.5.6** and visualized in a **Grafana 11.6.1** dashboard provisioned automatically on first boot. The backend emits structured JSON via **zerolog**; Promtail extracts `level`, `request_id`, and timestamp as Loki labels for fast log-line filtering.

---

## Numbers at a Glance

| Metric | Value |
| --- | --- |
| REST endpoints | 22 across 6 domains |
| Database tables | 9 (+2 migration scripts) |
| DB indexes | 13 (PK, UNIQUE, FK-support) |
| DB connection pool | 25 max open · 5 idle · 5-min lifetime |
| Redis cache TTL — ACL & team | 30 min |
| Redis cache TTL — asset metadata | 1 h |
| JWT expiry | 24 h |
| Import goroutine pool | 5 workers |
| Import progress flush cadence | every 500 rows or 2 s |
| Docker image size | ~30 MB (multi-stage, alpine base) |

---

## Tech Stack

| Layer | Technology |
| --- | --- |
| Backend | Go 1.26.2 · Gin 1.12.0 |
| Database | MySQL 8.4 · raw `database/sql` (no ORM) |
| Cache | Redis 7 · graceful noop fallback |
| Message queue | RabbitMQ 3 · graceful noop fallback |
| Auth | JWT HS256 · bcrypt |
| Frontend | React 19 · Vite 8 · TanStack Query v5 · Zustand v5 · Tailwind v4 |
| Observability | zerolog · Promtail · Loki · Grafana |
| Testing | `go-sqlmock` (no live DB required) |

---

## Quick Start

**Full stack via Docker (recommended):**

```bash
cp .env.example .env          # set JWT_SECRET at minimum
docker compose up -d
# API    → :8080
# App    → :3001
# Grafana → :3000
```

**Local development:**

```bash
# Backend
cd Backend
cp .env.example .env          # DB_HOST, JWT_SECRET required
go mod tidy
go run ./cmd                  # → :8080

# Frontend
cd Frontend/GoProject
npm install
npm run dev                   # → :5173
```

**Run tests (no live DB needed):**

```bash
cd Backend && go test ./...
```

---

## Backend Layer Structure

```text
Backend/
├── cmd/               # Entry point
├── internal/
│   ├── handler/       # HTTP layer  (7 handlers)
│   ├── service/       # Business logic + unit tests
│   ├── repository/    # SQL queries  (raw database/sql)
│   ├── model/         # Domain structs
│   ├── middleware/    # Auth · Rate limit · CORS · Logger
│   ├── cache/         # Redis client wrapper
│   ├── config/        # App + DB + CORS config
│   └── consumer/      # RabbitMQ event consumers
└── rabbitmq/          # Queue / exchange setup
```

---

## Known Limitations

| Area | Current State | Path Forward |
| --- | --- | --- |
| Token revocation | In-memory JTI map — resets on restart | Redis-backed blocklist with TTL |
| Rate limiting | Single-process in-memory | Redis-based limiter for multi-instance |
| Folder share inheritance | Point-in-time: notes added after sharing are not covered | Evaluate on-write vs. on-read resolution |
| Import progress | HTTP polling every 2 s | Server-Sent Events for true push |
| Pagination | All list endpoints return full result sets | `limit` / `offset` or cursor-based |
| Schema migrations | Manual `.sql` files | `golang-migrate` or Atlas |
