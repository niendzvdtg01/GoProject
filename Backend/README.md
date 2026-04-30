# Backend README

Backend hien tai la REST API viet bang Go + Gin, dung MySQL de luu user va JWT de xac thuc request.

## Muc tieu hien tai

Stage 1 tap trung vao Identity & Teams:

- Dang ky user voi role co dinh luc tao: `manager` hoac `member`.
- Dang nhap bang email/password.
- Tao JWT token cho session.
- Logout bang cach revoke token hien tai trong bo nho server.
- Lay danh sach users, chi cho phep role `manager`.

## Cau truc chinh

```text
Backend/
  api/MainRouting.go                  # Khai bao route API
  cmd/main.go                         # Entry point, load env, connect DB, start server
  internal/config/DbConfig.go         # Doc config database
  internal/handler/UserHandler.go     # HTTP handler cho auth/user
  internal/middleware/auth.go         # JWT middleware, role guard, token revoke
  internal/model/Users.go             # User model va role constants
  internal/respository/Database.go    # MySQL connection
  internal/respository/UserRepository.go
  internal/service/AuthService.go     # Login/logout logic
  internal/service/UserService.go     # Register user logic
  package/dtorequest/                 # Request DTO
  package/utils/vlidation.go          # Custom validator helpers
```

## Luong khoi dong

`cmd/main.go` thuc hien cac buoc:

1. Load `.env`.
2. Tao database config tu `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`.
3. Connect MySQL.
4. Tao `UserRepository`.
5. Doc `JWT_SECRET`.
6. Tao `AuthMiddleware`, `AuthService`, `UserService`.
7. Dang ky validators.
8. Setup router va chay server tai port `8080`.

## Bien moi truong

File `Backend/.env` can co:

```env
DB_HOST=localhost
DB_PORT=3311
DB_USERNAME=user
DB_PASSWORD=1234
DB_NAME=miniproject_database
JWT_SECRET=change-this-secret
```

Luu y: `DbConfig.go` hien dang hard-code port `3311`, chua doc `DB_PORT` tu `.env`.

## Database

Bang `users` can ton tai trong MySQL:

```sql
CREATE TABLE users (
  user_id VARCHAR(36) PRIMARY KEY,
  username VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role ENUM('manager', 'member') NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Y nghia field:

- `user_id`: UUID sinh tu server.
- `username`: ten hien thi.
- `email`: duy nhat, dung de login.
- `password_hash`: bcrypt hash, khong luu raw password.
- `role`: chi chap nhan `manager` hoac `member`.
- `created_at`: thoi diem tao user.

## API hien tai

Base URL:

```text
http://localhost:8080/api
```

### Register

```http
POST /api/users/register
Content-Type: application/json
```

Body:

```json
{
  "username": "manager01",
  "email": "manager@example.com",
  "password": "password123",
  "role": "manager"
}
```

Response thanh cong:

```json
{
  "token": "jwt-token",
  "user": {
    "userId": "uuid",
    "username": "manager01",
    "email": "manager@example.com",
    "role": "manager",
    "created_at": "0001-01-01T00:00:00Z"
  }
}
```

### Login

```http
POST /api/auth/login
Content-Type: application/json
```

Body:

```json
{
  "email": "manager@example.com",
  "password": "password123"
}
```

Response thanh cong:

```json
{
  "token": "jwt-token",
  "user": {
    "userId": "uuid",
    "username": "manager01",
    "email": "manager@example.com",
    "role": "manager",
    "created_at": "2026-04-30T10:00:00Z"
  }
}
```

### Logout

```http
POST /api/auth/logout
Authorization: Bearer <jwt-token>
```

Response thanh cong:

```json
{
  "message": "logout successful"
}
```

### List Users

Chi role `manager` duoc goi endpoint nay.

```http
GET /api/users
Authorization: Bearer <manager-jwt-token>
```

Response thanh cong:

```json
{
  "users": [
    {
      "userId": "uuid",
      "username": "manager01",
      "email": "manager@example.com",
      "role": "manager",
      "created_at": "2026-04-30T10:00:00Z"
    }
  ]
}
```

## Auth logic

- Password duoc hash bang `bcrypt`.
- JWT dung `HS256`.
- Token co `user_id`, `username`, `role`, `jti`, `issued_at`, `expires_at`.
- `AuthRequired()` doc header `Authorization: Bearer <token>`, validate token va set context:
  - `user_id`
  - `username`
  - `role`
- `RoleRequired("manager")` kiem tra role trong context.
- Logout dua `jti` cua token vao blacklist in-memory cho den khi token het han.

Luu y: blacklist logout hien chi nam trong memory. Neu server restart, token da logout truoc do se khong con nam trong blacklist.

## Chay project

Tu thu muc `Backend`:

```bash
go mod tidy
go run ./cmd
```

Test compile:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```

## Cac diem can hoan thien tiep

- Doc `DB_PORT` tu `.env` thay vi hard-code.
- Them migration thay vi tao bang thu cong.
- Luu revoked token vao Redis/database neu can logout co hieu luc sau khi restart server.
- Them refresh token neu can session dai han.
- Them test cho service va middleware.
