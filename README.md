# Microservices Challenge: User, Team & Asset Management

## Overview

This project is a Go-based backend system for managing users, teams, and digital assets. The current implementation starts with Stage 1: identity and user management. The long-term architecture is designed to grow into a simple microservices system where each domain owns its own business logic and data.

The system will support two user roles:

- `manager`: can manage users and teams.
- `member`: can access the system and manage assigned assets.

## Planned Architecture

The project can be separated into these services:

```text
Client
  |
  v
API Gateway / HTTP Router
  |
  +-- Auth & User Service
  +-- Team Service
  +-- Asset Service
  +-- Notification / Event Worker
```

### Auth & User Service

This service is responsible for identity:

- User registration
- Login and logout
- Password hashing
- JWT token generation
- Role validation
- User listing for managers

Current backend code already implements most of this service using Go, Gin, MySQL, bcrypt, and JWT.

### Team Service

This service will manage team-related behavior:

- Create teams
- Add or remove members
- Assign managers
- List teams by user
- Validate team-level permissions

Managers should be able to create and manage teams. Members should only see teams they belong to.

### Asset Service

This service will manage digital assets:

- Folders
- Notes
- Ownership
- Sharing rules
- Permission checks

Assets should support granular access control, for example read-only access, edit access, or owner access.

### Event Worker

For a simple first version, services can communicate through REST APIs. Later, an event worker or message queue can be added for asynchronous tasks such as:

- Audit logs
- Notifications
- Asset sharing events
- Team membership changes

## Current Tech Stack

- Language: Go
- HTTP Framework: Gin
- Database: MySQL
- Authentication: JWT
- Password Hashing: bcrypt
- Configuration: `.env`

## Why This Stack

Go is a good fit for backend services because it is simple, fast, and has strong support for concurrent workloads. Gin is lightweight and easy to use for REST APIs, making it practical for the early stages of the project.

MySQL is used because the current data model is relational: users, teams, memberships, folders, notes, and permissions all have clear relationships. SQL constraints such as unique email and foreign keys can help protect data consistency.

JWT is used for stateless authentication. It keeps the API simple because each request can carry its own identity and role claims. The trade-off is that logout and token revocation require extra handling, such as an in-memory or Redis-backed blacklist.

## Current Backend Structure

```text
Backend/
  api/                 # Route definitions
  cmd/                 # Application entry point
  internal/config/     # Environment and DB config
  internal/handler/    # HTTP handlers
  internal/middleware/ # JWT auth, role guards, request middleware
  internal/model/      # Domain models
  internal/respository/# Database access layer
  internal/service/    # Business logic
  package/             # Shared DTOs and utilities
```

## Main API Flow

1. A user registers with username, email, password, and role.
2. The password is hashed using bcrypt before saving.
3. The user logs in with email and password.
4. The server returns a JWT token.
5. Protected endpoints require `Authorization: Bearer <token>`.
6. Middleware validates the token and checks role permissions.
7. Managers can access user-management endpoints.

## Development Roadmap

### Stage 1: Identity Foundation

- User registration
- Login/logout
- JWT authentication
- Role-based authorization
- User list endpoint for managers

### Stage 2: Team Management

- Team model
- Team membership model
- Manager-only team creation
- Add/remove members
- Team permission checks

### Stage 3: Asset Management

- Folder and note models
- Asset ownership
- Sharing and access control
- Audit logs or event-driven updates

## Future Improvement

The first major bottleneck will be authentication and authorization across multiple services. If the system grows, each service should not duplicate complex permission logic. A future improvement is to introduce a centralized authorization layer or policy service.

For example, the system could use:

- Redis for shared token blacklist and session metadata
- A policy engine for access control rules
- Message queue for asynchronous service communication
- API Gateway for centralized authentication before routing requests to services

This would keep each service focused on its own domain while still enforcing consistent security rules across the whole platform.

## Running The Backend

From the `Backend` directory:

```bash
go run ./cmd
```

Run compile checks:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```
