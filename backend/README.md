# Backend

The backend provides the REST API responsible for onboarding applications into Keycloak. It manages application registration, provisioning workflows, template management, Keycloak integration, provisioning jobs, and platform configuration.

---

# Features

- RESTful API
- Application Management
- Keycloak Client Provisioning
- Client Role Management
- Protocol Mapper Management
- Client Scope Assignment
- Provisioning Job Tracking
- Template Management
- Platform Settings
- PostgreSQL Persistence
- Request Logging
- Recovery Middleware
- Request ID Middleware
- CORS Support
- Optional API Authentication

---

# Technology Stack

| Component        | Technology            |
| ---------------- | --------------------- |
| Language         | Go 1.25+              |
| HTTP Framework   | Gin                   |
| Database         | PostgreSQL            |
| Driver           | pgx                   |
| Configuration    | Environment Variables |
| Containerization | Docker                |
| Migrations       | golang-migrate        |

---

# Directory Structure

```text
backend/

├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│
│   ├── applications/
│   │   ├── dto.go
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   ├── keycloak/
│   │   ├── client.go
│   │   ├── clients.go
│   │   ├── roles.go
│   │   ├── scopes.go
│   │   ├── mappers.go
│   │   ├── secrets.go
│   │   └── token.go
│   │
│   ├── provisioning/
│   │   ├── dto.go
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── provisioner.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   ├── templates/
│   │   ├── dto.go
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── seed.go
│   │   └── service.go
│   │
│   ├── settings/
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── requestid.go
│   │
│   ├── database/
│   ├── config/
│   ├── response/
│   └── server/
│
├── migrations/
│
├── Dockerfile
│
├── go.mod
└── go.sum
```

---

# Local Development

## Prerequisites

Install:

- Go 1.25+
- PostgreSQL
- Docker
- Docker Compose

Clone the project.

```bash
git clone <repository-url>

cd keycloak-onboarder/backend
```

---

# Install Dependencies

```bash
go mod download
```

or

```bash
go mod tidy
```

---

# Running the API

```bash
go run ./cmd/api
```

Default server

```
http://localhost:9000
```

Health endpoint

```
GET /api/v1/health
```

---

# Building

Build a binary.

```bash
go build -o api ./cmd/api
```

Run.

```bash
./api
```

---

# Environment Variables

Example `.env`

```env
APP_ENV=development

SERVER_PORT=9000

DATABASE_URL=postgres://postgres:postgres@localhost:5432/onboarder?sslmode=disable

KEYCLOAK_BASE_URL=http://localhost:8080
KEYCLOAK_REALM=master
KEYCLOAK_ADMIN_CLIENT_ID=onboarder-admin
KEYCLOAK_ADMIN_CLIENT_SECRET=

ALLOWED_ORIGINS=http://localhost:3000

API_AUTH_TOKEN=
```

---

# Docker

Build backend.

```bash
docker compose build backend
```

Run backend.

```bash
docker compose up backend
```

View logs.

```bash
docker compose logs -f backend
```

Open a shell.

```bash
docker compose exec backend sh
```

---

# Database

Run migrations.

```bash
docker compose up migrate
```

Reset database.

```bash
docker compose down -v
```

Restart.

```bash
docker compose up
```

---

# Middleware

The backend includes the following middleware.

## Request ID

Generates a unique request ID for every request.

Response Header

```
X-Request-ID
```

---

## Logger

Logs every request.

Example

```
GET /api/v1/applications

Status 200

Duration 15ms

Request ID
```

---

## Recovery

Recovers from panics.

Returns

```json
{
    "error":"internal server error"
}
```

---

## CORS

Supports configurable origins.

```
ALLOWED