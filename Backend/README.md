# Backend

Go REST API using Gin, MySQL, bcrypt, and JWT.

## Architecture

Strict three-layer pattern: **Handler → Service → Repository → MySQL**

```
cmd/main.go              # Entry point: wires all layers, starts on :8080
api/MainRouting.go       # Route registration
internal/
  config/                # DB connection pool, CORS config
  handler/               # HTTP handlers (parse → service → respond)
    AuthHandler.go
    UserHandler.go
    TeamHandler.go
    FolderHandler.go
    NoteHandler.go
    SharingHandler.go
    ImportHandler.go
  middleware/            # JWT auth, role check, rate limit, CORS, request ID
  model/                 # User, Team, TeamMember, Folder, Note, Permission
  repository/            # Raw SQL via database/sql — no ORM
  service/               # Business logic
    AuthService.go
    UserService.go       # includes bulk CSV import
    TeamManagementService.go
    FolderService.go
    NoteService.go
    SharingService.go
package/
  dtorequest/            # Input DTOs with validator struct tags
  utils/                 # Custom validator registration
```

## Environment Setup

Create `Backend/.env`:

```env
DB_HOST=localhost
DB_PORT=3311
DB_USERNAME=user
DB_PASSWORD=1234
DB_NAME=miniproject_database
JWT_SECRET=your_secret_here
```

Server panics on startup if `JWT_SECRET` is missing.

## Run

```bash
# From Backend/
go mod tidy
go run ./cmd
```

## Database Schema

```sql
CREATE TABLE users (
  user_id      VARCHAR(36)  PRIMARY KEY,
  username     VARCHAR(100) NOT NULL UNIQUE,
  email        VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role         VARCHAR(20)  NOT NULL DEFAULT 'member',
  created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teams (
  team_id    INT          PRIMARY KEY AUTO_INCREMENT,
  team_name  VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE team_members (
  team_id   INT         NOT NULL,
  user_id   VARCHAR(36) NOT NULL,
  role      VARCHAR(20) NOT NULL DEFAULT 'MEMBER',
  joined_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (team_id, user_id),
  FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE folders (
  id         INT          PRIMARY KEY AUTO_INCREMENT,
  owner_id   VARCHAR(36)  NOT NULL,
  name       VARCHAR(255) NOT NULL,
  created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE notes (
  id         INT          PRIMARY KEY AUTO_INCREMENT,
  folder_id  INT          NOT NULL,
  owner_id   VARCHAR(36)  NOT NULL,
  title      VARCHAR(255) NOT NULL,
  content    TEXT         NOT NULL,
  created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
  FOREIGN KEY (owner_id)  REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE permissions (
  id              INT         PRIMARY KEY AUTO_INCREMENT,
  asset_type      VARCHAR(10) NOT NULL,   -- 'folder' or 'note'
  asset_id        INT         NOT NULL,
  user_id         VARCHAR(36) NOT NULL,
  permission_type VARCHAR(10) NOT NULL,   -- 'read' or 'write'
  granted_by      VARCHAR(36) NOT NULL,
  created_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_perm (asset_type, asset_id, user_id)
);
```

## API Reference

Base URL: `http://localhost:8080/api`

All protected endpoints require `Authorization: Bearer <token>`.

---

### Auth

#### `POST /auth/login`
```json
// Request
{ "email": "user@example.com", "password": "password123" }

// Response 200
{ "token": "<jwt>", "user": { "user_id": "...", "username": "...", "email": "...", "role": "...", "created_at": "..." } }
```

#### `POST /auth/logout` 🔒
Revokes the current token (in-memory, cleared on restart).
```json
// Response 200
{ "message": "logout successful" }
```

---

### Users

#### `POST /users/register`
```json
// Request
{ "username": "alice", "email": "alice@example.com", "password": "secret123", "role": "member" }

// Response 201 — same shape as login response
```

#### `GET /users` 🔒
Returns all registered users.
```json
{ "users": [ { "user_id": "...", "username": "...", "email": "...", "role": "...", "created_at": "..." } ] }
```

#### `POST /users/import` 🔒
Upload a `.csv` file with columns `username,email,password[,role]`. Role defaults to `member`.
Processed concurrently by a 5-worker goroutine pool.

```
// multipart/form-data, field name: "file"

// Response 200
{ "succeeded": 48, "failed": 2, "errors": [ { "email": "x@y.com", "error": "email already exists" } ] }
```

---

### Teams

#### `GET /teams` 🔒
Returns all teams the current user belongs to.
```json
{ "teams": [ { "team_id": 1, "team_name": "backend", "created_at": "...", "updated_at": "..." } ] }
```

#### `POST /teams` 🔒 Manager only
```json
// Request
{ "teamName": "backend", "members": [ { "username": "bob", "role": "MEMBER" } ] }

// Response 201
{ "teamID": 1 }
```

#### `POST /teams/:teamName/members` 🔒 Manager only
```json
// Request
{ "username": "carol", "role": "MEMBER" }
```

#### `DELETE /teams/:teamName/members/:memberName` 🔒 Manager only

#### `DELETE /teams/:teamName` 🔒 Manager only (OWNER role in team required)

---

### Folders

#### `GET /folders` 🔒
List all folders owned by the current user.

#### `POST /folders` 🔒
```json
{ "name": "My Folder" }
```

#### `GET /folders/:id` 🔒
Owner, explicitly shared users, or managers of same team can read.

#### `PUT /folders/:id` 🔒
```json
{ "name": "Renamed Folder" }
```
Owner or users with write permission only.

#### `DELETE /folders/:id` 🔒
Owner only.

---

### Notes

#### `GET /folders/:id/notes` 🔒
List notes in a folder (requires read access to folder).

#### `POST /folders/:id/notes` 🔒
```json
{ "title": "Meeting notes", "content": "Discussed roadmap..." }
```

#### `GET /notes/:id` 🔒

#### `PUT /notes/:id` 🔒
```json
{ "title": "Updated title", "content": "Updated content" }
```

#### `DELETE /notes/:id` 🔒

---

### Sharing

#### `POST /share` 🔒
Share a folder or note with a user by email.
```json
{
  "email": "bob@example.com",
  "folder_id": 1,
  "permission_type": "read"
}
```
`folder_id` or `note_id` — use one. Sharing a folder automatically shares all its notes (inheritance).

#### `DELETE /share` 🔒
Revoke access.
```json
{
  "email": "bob@example.com",
  "folder_id": 1
}
```

---

## Access Control Rules

| Actor | Action | Condition |
|---|---|---|
| Owner | Read / Write / Delete | Always |
| Shared (write) | Read / Write | Explicit permission |
| Shared (read) | Read only | Explicit permission |
| Manager (same team) | Read only | Implicit — no explicit share needed |
| Manager + shared (write) | Read / Write | Explicit write permission |
| Other users | None | Denied (403) |

## Key Patterns

- **Adding an endpoint**: DTO → Repository method → Service method → Handler method → Route in `MainRouting.go`
- **Validation**: `c.ShouldBindJSON(&req)` + `binding:"required,..."` struct tags
- **Context**: all repository methods accept `context.Context` as first argument
- **Errors**: MySQL error code 1062 → `ErrEmailAlreadyExists`; `sql.ErrNoRows` → domain `ErrNotFound` variants

## Known Limitations

- Token revocation is in-memory — revoked tokens re-activate on server restart
- No refresh token
- No database migration framework — create tables manually using the SQL above
- No automated tests
